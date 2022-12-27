package message

import (
	"encoding/json"

	"github.com/markusylisiurunen/go-opinionatedevents"
)

type createOrderTaskDataCustomer struct {
	Address string `json:"address"`
	Name    string `json:"name"`
}

type createOrderTaskDataProduct struct {
	Price    int    `json:"price"`
	Quantity int    `json:"quantity"`
	UUID     string `json:"uuid"`
}

type createOrderTaskData struct {
	Customer createOrderTaskDataCustomer  `json:"customer"`
	Products []createOrderTaskDataProduct `json:"products"`
}

type TaskCreateOrder struct {
	message
	Task createOrderTaskData `json:"task"`
}

type TaskCreateOrderParamsCustomer struct {
	Address string
	Name    string
}

type TaskCreateOrderParamsProduct struct {
	Price    int
	Quantity int
	UUID     string
}

type TaskCreateOrderParams struct {
	Customer TaskCreateOrderParamsCustomer
	Products []TaskCreateOrderParamsProduct
}

func NewTaskCreateOrder(args *TaskCreateOrderParams, previous hasRollbackStack) *TaskCreateOrder {
	if previous == nil {
		previous = rollbackStack{}
	}
	products := []createOrderTaskDataProduct{}
	for _, product := range args.Products {
		products = append(products, createOrderTaskDataProduct(product))
	}
	return &TaskCreateOrder{
		message: message{
			Meta:          map[string]string{},
			RollbackStack: previous.getRollbackStack(),
		},
		Task: createOrderTaskData{
			Customer: createOrderTaskDataCustomer{
				Address: args.Customer.Address,
				Name:    args.Customer.Name,
			},
			Products: products,
		},
	}
}

func (m *TaskCreateOrder) Name() string {
	return "tasks.create_order"
}

func (m *TaskCreateOrder) MarshalPayload() ([]byte, error) {
	return json.Marshal(m)
}

func (m *TaskCreateOrder) UnmarshalPayload(data []byte) error {
	return json.Unmarshal(data, m)
}

func (m *TaskCreateOrder) GetOpinionated() *opinionatedevents.Message {
	v, err := opinionatedevents.NewMessage(m.Name(), m)
	if err != nil {
		panic(err)
	}
	return v
}
