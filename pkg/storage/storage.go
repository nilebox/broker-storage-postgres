package storage

import (
	"database/sql"
	"encoding/json"

	"context"
	"time"

	"github.com/lib/pq"
	brokerstorage "github.com/nilebox/broker-server/pkg/stateful/storage"
	"github.com/nilebox/broker-storage-postgres/pkg/storage/db"
	"github.com/pkg/errors"
)

type postgresStorage struct {
	db            *sql.DB
	ctx           context.Context
	leaseDuration time.Duration
}

type instanceRow struct {
	InstanceId string
	ServiceId  string
	PlanId     string
	Parameters string
	Outputs    string
	State      string
	Error      string
	Created    time.Time
	Modified   time.Time
}

const (
	QueryGetInstance = "SELECT instance_id, service_id, plan_id, parameters, outputs, state, error " +
		"FROM instance WHERE instance_id = $1"
	// TODO don't set outputs there
	QueryInsertInstance = "INSERT INTO instance (instance_id, service_id, plan_id, parameters, outputs, state, created, modified, error) " +
		"VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)"
	// TODO don't set outputs there
	QueryUpdateInstance = "UPDATE instance SET plan_id = $1, parameters = $2, outputs = $3, state = $4, modified = $5, error = $6 " +
		"WHERE instance_id = $7"
	QueryUpdateInstanceState = "UPDATE instance SET error = $1, state = $2, modified = $3 " +
		"WHERE instance_id = :instanceId"
	QueryUpdateInstanceOutputs = "UPDATE instance SET outputs = $1 " +
		"WHERE instance_id = $2"
	QueryExtendLease = "UPDATE instance SET modified = $1 " +
		"WHERE instance_id = ANY($2)"
	QueryLeaseAbandoned = "UPDATE instance SET modified = $1 " +
		"WHERE id IN (" +
		"   SELECT id FROM instance " +
		"   WHERE state = ANY($2) AND modified < $3 " +
		"   FOR UPDATE SKIP LOCKED " +
		"   LIMIT $4 " +
		") " +
		"RETURNING instance_id, service_id, plan_id, parameters, outputs, state, error"
)

var (
	inProgressStates = []string{
		string(brokerstorage.InstanceStateCreateInProgress),
		string(brokerstorage.InstanceStateCreateInProgress),
		string(brokerstorage.InstanceStateCreateInProgress),
	}
)

func NewPostgresStorage(ctx context.Context, config *db.PostgresConfig) (*postgresStorage, error) {
	conn, err := db.OpenConnection(config)
	if err != nil {
		return nil, err
	}
	err = db.RunMigrations(ctx, conn)
	if err != nil {
		closeErr := conn.Close()
		if closeErr != nil {
			return nil, closeErr
		}
		return nil, err
	}

	return &postgresStorage{
		ctx:           ctx,
		db:            conn,
		leaseDuration: time.Minute * 5,
	}, nil
}

func (s *postgresStorage) Close() error {
	return s.db.Close()
}

func (s *postgresStorage) CreateInstance(instanceSpec *brokerstorage.InstanceSpec) error {
	instance := &brokerstorage.InstanceRecord{
		Spec: *instanceSpec,
	}
	row, err := instanceRecordToRow(instance)
	row.State = string(brokerstorage.InstanceStateCreateInProgress)
	row.Created = time.Now()
	row.Modified = row.Created
	if err != nil {
		return err
	}
	_, err = db.InTransaction(s.db, func(tx *sql.Tx) (result interface{}, returnErr error) {
		_, err := tx.ExecContext(s.ctx, QueryInsertInstance,
			row.InstanceId, row.ServiceId, row.PlanId, row.Parameters, row.Outputs, row.State, row.Created, row.Modified, row.Error)
		if err != nil {
			return nil, errors.Wrap(err, "failed to insert instance")
		}
		return nil, nil
	})
	return err
}

func (s *postgresStorage) UpdateInstance(instanceSpec *brokerstorage.InstanceSpec) error {
	instanceToUpdate, err := s.GetInstance(instanceSpec.InstanceId)
	if err != nil {
		return errors.Wrap(err, "failed to fetch the instance")
	}
	instanceToUpdate.Spec.PlanId = instanceSpec.PlanId
	instanceToUpdate.Spec.Parameters = instanceSpec.Parameters
	instanceToUpdate.Spec.Outputs = instanceSpec.Outputs

	row, err := instanceRecordToRow(instanceToUpdate)
	row.State = string(brokerstorage.InstanceStateUpdateInProgress)
	row.Modified = time.Now()
	if err != nil {
		return err
	}
	_, err = db.InTransaction(s.db, func(tx *sql.Tx) (result interface{}, returnErr error) {
		_, err := tx.ExecContext(s.ctx, QueryUpdateInstance,
			row.PlanId, row.Parameters, row.Outputs, row.State, row.Modified, row.Error,
			row.InstanceId)
		if err != nil {
			return nil, errors.Wrap(err, "failed to update instance")
		}
		return nil, nil
	})
	return err
}

func (s *postgresStorage) DeleteInstance(instanceId string) error {
	// TODO implement
	panic("NotImplemented")
}

func (s *postgresStorage) UpdateInstanceState(instanceId string, state brokerstorage.InstanceState, errorMessage string) error {
	_, err := db.InTransaction(s.db, func(tx *sql.Tx) (result interface{}, returnErr error) {
		_, err := tx.ExecContext(s.ctx, QueryUpdateInstanceState,
			errorMessage, string(state), time.Now(),
			instanceId)
		if err != nil {
			return nil, errors.Wrap(err, "failed to update instance status")
		}
		return nil, nil
	})
	return err
}

func (s *postgresStorage) UpdateInstanceOutputs(instanceId string, outputs json.RawMessage) error {
	outputsStr, err := rawMessageToString(outputs)
	if err != nil {
		return err
	}
	_, err = db.InTransaction(s.db, func(tx *sql.Tx) (result interface{}, returnErr error) {
		_, err := tx.ExecContext(s.ctx, QueryUpdateInstanceOutputs,
			outputsStr,
			instanceId)
		if err != nil {
			return nil, errors.Wrap(err, "failed to update instance outputs")
		}
		return nil, nil
	})
	return err
}

func (s *postgresStorage) GetInstance(instanceId string) (*brokerstorage.InstanceRecord, error) {
	rows, err := s.db.QueryContext(s.ctx, QueryGetInstance, instanceId)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get instance %s", instanceId)
	}
	var returnErr error
	defer db.Close(rows, &returnErr)

	row, err := getSingleInstanceRow(rows)
	if err != nil {
		return nil, err
	}
	record, returnErr := instanceRowToRecord(row)
	return record, returnErr
}

func (s *postgresStorage) ExtendLease(instanceIds []string) error {
	_, err := db.InTransaction(s.db, func(tx *sql.Tx) (result interface{}, returnErr error) {
		_, err := tx.ExecContext(s.ctx, QueryExtendLease,
			time.Now(),
			pq.Array(instanceIds))
		if err != nil {
			return nil, errors.Wrap(err, "failed to update instance status")
		}
		return nil, nil
	})
	return err
}

func (s *postgresStorage) LeaseAbandonedInstances(maxBatchSize uint32) ([]*brokerstorage.InstanceRecord, error) {
	now := time.Now()
	expire := now.Add(-s.leaseDuration)
	rows, err := s.db.QueryContext(s.ctx, QueryLeaseAbandoned, time.Now(), inProgressStates, expire, maxBatchSize)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to lease instances")
	}
	var returnErr error
	defer db.Close(rows, &returnErr)

	instanceRows, err := getInstanceRows(rows)
	if err != nil {
		return nil, returnErr
	}
	instanceRecords := make([]*brokerstorage.InstanceRecord, 0, len(instanceRows))
	for i, instanceRow := range instanceRows {
		instanceRecord, err := instanceRowToRecord(instanceRow)
		if err != nil {
			return nil, returnErr
		}
		instanceRecords[i] = instanceRecord
	}
	return instanceRecords, returnErr
}

func getInstanceRows(rows *sql.Rows) ([]*instanceRow, error) {
	instanceRows := []*instanceRow{}
	for rows.Next() {
		instanceRow, err := scanInstanceRow(rows)
		if err != nil {
			return nil, err
		}
		instanceRows = append(instanceRows, instanceRow)
	}
	return instanceRows, nil
}

func getSingleInstanceRow(rows *sql.Rows) (*instanceRow, error) {
	if rows.Next() {
		return scanInstanceRow(rows)
	}
	return nil, nil
}

func scanInstanceRow(rows *sql.Rows) (*instanceRow, error) {
	var row instanceRow
	if err := rows.Scan(&row.InstanceId, &row.ServiceId, &row.PlanId, &row.Parameters, &row.Outputs, &row.State, &row.Error); err != nil {
		return nil, errors.Wrap(err, "failed to scan row")
	}
	return &row, nil
}

func instanceRowToRecord(row *instanceRow) (*brokerstorage.InstanceRecord, error) {
	parameters, err := rawMessageFromString(row.Parameters)
	if err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal parameters")
	}
	outputs, err := rawMessageFromString(row.Outputs)
	if err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal outputs")
	}
	record := &brokerstorage.InstanceRecord{
		Spec: brokerstorage.InstanceSpec{
			InstanceId: row.InstanceId,
			ServiceId:  row.ServiceId,
			PlanId:     row.PlanId,
			Parameters: parameters,
			Outputs:    outputs,
		},
		State: brokerstorage.InstanceState(row.State),
		Error: row.Error,
	}
	return record, nil
}

func instanceRecordToRow(record *brokerstorage.InstanceRecord) (*instanceRow, error) {
	parameters, err := rawMessageToString(record.Spec.Parameters)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal parameters")
	}
	outputs, err := rawMessageToString(record.Spec.Outputs)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal outputs")
	}
	row := &instanceRow{
		InstanceId: record.Spec.InstanceId,
		ServiceId:  record.Spec.ServiceId,
		PlanId:     record.Spec.PlanId,
		Parameters: parameters,
		Outputs:    outputs,
		State:      string(record.State),
		Error:      record.Error,
	}
	return row, nil
}

func rawMessageToString(message json.RawMessage) (string, error) {
	if message == nil {
		return "", nil
	}
	bytes, err := json.Marshal(message)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func rawMessageFromString(str string) (json.RawMessage, error) {
	if str == "" {
		return nil, nil
	}
	var message json.RawMessage
	err := json.Unmarshal([]byte(str), &message)
	if err != nil {
		return nil, err
	}
	return message, nil
}
