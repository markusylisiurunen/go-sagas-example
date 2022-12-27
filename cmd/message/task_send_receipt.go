package message

import (
	"encoding/json"

	"github.com/markusylisiurunen/go-opinionatedevents"
)

type sendReceiptTaskDataCustomer struct {
	Address string `json:"address"`
	Name    string `json:"name"`
}

type sendReceiptTaskDataProduct struct {
	Price    int    `json:"price"`
	Quantity int    `json:"quantity"`
	UUID     string `json:"uuid"`
}

type sendReceiptTaskData struct {
	Customer sendReceiptTaskDataCustomer  `json:"customer"`
	Products []sendReceiptTaskDataProduct `json:"products"`
}

type TaskSendReceipt struct {
	message
	Task sendReceiptTaskData `json:"task"`
}

type TaskSendReceiptParamsCustomer struct {
	Address string
	Name    string
}

type TaskSendReceiptParamsProduct struct {
	Price    int
	Quantity int
	UUID     string
}

type TaskSendReceiptParams struct {
	Customer TaskSendReceiptParamsCustomer
	Products []TaskSendReceiptParamsProduct
}

func NewTaskSendReceipt(args *TaskSendReceiptParams, previous hasRollbackStack) *TaskSendReceipt {
	if previous == nil {
		previous = rollbackStack{}
	}
	products := []sendReceiptTaskDataProduct{}
	for _, product := range args.Products {
		products = append(products, sendReceiptTaskDataProduct(product))
	}
	return &TaskSendReceipt{
		message: message{
			Meta:          map[string]string{},
			RollbackStack: previous.getRollbackStack(),
		},
		Task: sendReceiptTaskData{
			Customer: sendReceiptTaskDataCustomer{
				Address: args.Customer.Address,
				Name:    args.Customer.Name,
			},
			Products: products,
		},
	}
}

func (m *TaskSendReceipt) Name() string {
	return "tasks.send_receipt"
}

func (m *TaskSendReceipt) MarshalPayload() ([]byte, error) {
	return json.Marshal(m)
}

func (m *TaskSendReceipt) UnmarshalPayload(data []byte) error {
	return json.Unmarshal(data, m)
}

func (m *TaskSendReceipt) GetOpinionated() *opinionatedevents.Message {
	v, err := opinionatedevents.NewMessage(m.Name(), m)
	if err != nil {
		panic(err)
	}
	return v
}
