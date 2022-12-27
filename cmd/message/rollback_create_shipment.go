package message

import (
	"encoding/json"

	"github.com/markusylisiurunen/go-opinionatedevents"
)

type rollbackCreateShipmentTaskData struct {
	Shipment string `json:"shipment"`
}

type RollbackCreateShipment struct {
	message
	Task rollbackCreateShipmentTaskData `json:"task"`
}

type RollbackCreateShipmentParams struct {
	Shipment string
}

func NewRollbackCreateShipment(args *RollbackCreateShipmentParams) *RollbackCreateShipment {
	return &RollbackCreateShipment{
		message: message{
			Meta:          map[string]string{},
			RollbackStack: rollbackStack{},
		},
		Task: rollbackCreateShipmentTaskData{
			Shipment: args.Shipment,
		},
	}
}

func (m *RollbackCreateShipment) getRollbackStackItem() rollbackStackItem {
	return rollbackStackItem{
		Name: m.Name(),
		Meta: m.Meta,
		Data: map[string]any{
			"shipment": m.Task.Shipment,
		},
	}
}

func (m *RollbackCreateShipment) Name() string {
	return "tasks.rollback_create_shipment"
}

func (m *RollbackCreateShipment) MarshalPayload() ([]byte, error) {
	return json.Marshal(m)
}

func (m *RollbackCreateShipment) UnmarshalPayload(data []byte) error {
	return json.Unmarshal(data, m)
}

func (m *RollbackCreateShipment) GetOpinionated() *opinionatedevents.Message {
	v, err := opinionatedevents.NewMessage(m.Name(), m)
	if err != nil {
		panic(err)
	}
	return v
}
