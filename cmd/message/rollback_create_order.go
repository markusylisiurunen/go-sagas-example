package message

import (
	"encoding/json"

	"github.com/markusylisiurunen/go-opinionatedevents"
)

type rollbackCreateOrderTaskData struct {
	Order string `json:"order"`
}

type RollbackCreateOrder struct {
	message
	Task rollbackCreateOrderTaskData `json:"task"`
}

type RollbackCreateOrderParams struct {
	Order string
}

func NewRollbackCreateOrder(args *RollbackCreateOrderParams) *RollbackCreateOrder {
	return &RollbackCreateOrder{
		message: message{
			Meta:          map[string]string{},
			RollbackStack: rollbackStack{},
		},
		Task: rollbackCreateOrderTaskData{
			Order: args.Order,
		},
	}
}

func (m *RollbackCreateOrder) getRollbackStackItem() rollbackStackItem {
	return rollbackStackItem{
		Name: m.Name(),
		Meta: m.Meta,
		Data: map[string]any{
			"order": m.Task.Order,
		},
	}
}

func (m *RollbackCreateOrder) Name() string {
	return "tasks.rollback_create_order"
}

func (m *RollbackCreateOrder) MarshalPayload() ([]byte, error) {
	return json.Marshal(m)
}

func (m *RollbackCreateOrder) UnmarshalPayload(data []byte) error {
	return json.Unmarshal(data, m)
}

func (m *RollbackCreateOrder) GetOpinionated() *opinionatedevents.Message {
	v, err := opinionatedevents.NewMessage(m.Name(), m)
	if err != nil {
		panic(err)
	}
	return v
}
