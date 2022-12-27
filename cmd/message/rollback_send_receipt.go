package message

import (
	"encoding/json"

	"github.com/markusylisiurunen/go-opinionatedevents"
)

type rollbackSendReceiptTaskData struct {
	Receipt string `json:"receipt"`
}

type RollbackSendReceipt struct {
	message
	Task rollbackSendReceiptTaskData `json:"task"`
}

type RollbackSendReceiptParams struct {
	Receipt string
}

func NewRollbackSendReceipt(args *RollbackSendReceiptParams) *RollbackSendReceipt {
	return &RollbackSendReceipt{
		message: message{
			Meta:          map[string]string{},
			RollbackStack: rollbackStack{},
		},
		Task: rollbackSendReceiptTaskData{
			Receipt: args.Receipt,
		},
	}
}

func (m *RollbackSendReceipt) getRollbackStackItem() rollbackStackItem {
	return rollbackStackItem{
		Name: m.Name(),
		Meta: m.Meta,
		Data: map[string]any{
			"receipt": m.Task.Receipt,
		},
	}
}

func (m *RollbackSendReceipt) Name() string {
	return "tasks.rollback_send_receipt"
}

func (m *RollbackSendReceipt) MarshalPayload() ([]byte, error) {
	return json.Marshal(m)
}

func (m *RollbackSendReceipt) UnmarshalPayload(data []byte) error {
	return json.Unmarshal(data, m)
}

func (m *RollbackSendReceipt) GetOpinionated() *opinionatedevents.Message {
	v, err := opinionatedevents.NewMessage(m.Name(), m)
	if err != nil {
		panic(err)
	}
	return v
}
