package broker

import (
	"encoding/json"

	brokerapi "github.com/nilebox/broker-server/pkg/api"
	"go.uber.org/zap"

	"github.com/nilebox/broker-server/pkg/stateful/task"
)

const (
	ServiceId   = "uuid1"
	ServiceName = "my-service"
	PlanId      = "uuid2"
	PlanName    = "default"
)

type exampleBroker struct {
	log *zap.Logger
}

func NewExampleBroker(log *zap.Logger) (task.Broker, error) {
	return &exampleBroker{
		log: log,
	}, nil
}

func Catalog() *brokerapi.Catalog {
	catalog := brokerapi.Catalog{
		Services: []*brokerapi.Service{
			{
				ID:          ServiceId,
				Name:        ServiceName,
				Description: "Service description",
				Bindable:    true,
				Plans: []brokerapi.ServicePlan{
					{
						ID:          PlanId,
						Name:        PlanName,
						Description: "Plan description",
						//Schemas: &brokerapi.Schemas{
						//	Instance: brokerapi.InstanceSchema{
						//		Create: brokerapi.Schema{
						//			Parameters: c.schema,
						//		},
						//		Update: brokerapi.Schema{
						//			Parameters: c.schema,
						//		},
						//	},
						//},
					},
				},
				PlanUpdateable: false,
			},
		},
	}

	return &catalog
}

func (c *exampleBroker) CreateInstance(instanceId string, parameters json.RawMessage) (json.RawMessage, error) {
	return nil, nil
}

func (c *exampleBroker) UpdateInstance(instanceId string, parameters json.RawMessage) (json.RawMessage, error) {
	return nil, nil
}

func (c *exampleBroker) DeleteInstance(instanceId string, parameters json.RawMessage) error {
	return nil
}
