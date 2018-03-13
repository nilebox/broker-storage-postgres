package storage

import (
	"encoding/json"
	"github.com/nilebox/broker-server/pkg/stateful/retry"
	brokerstorage "github.com/nilebox/broker-server/pkg/stateful/storage"
)

type postgresStorage struct {
}

func NewPostgresStorage() retry.StorageWithLease {
	return &postgresStorage{}
}

func (s *postgresStorage) CreateInstance(instance *brokerstorage.InstanceRecord) error {
	return nil
}

func (s *postgresStorage) UpdateInstance(instanceId string, parameters json.RawMessage, state brokerstorage.InstanceState) error {
	return nil
}

func (s *postgresStorage) UpdateInstanceState(instanceId string, state brokerstorage.InstanceState, err string) error {
	return nil
}

func (s *postgresStorage) GetInstance(instanceId string) (*brokerstorage.InstanceRecord, error) {
	return nil, nil
}

func (s *postgresStorage) ExtendLease(instances []*brokerstorage.InstanceRecord) error {
	return nil
}

func (s *postgresStorage) LeaseAbandonedInstances(maxBatchSize uint32) []*brokerstorage.InstanceRecord {
	return nil
}
