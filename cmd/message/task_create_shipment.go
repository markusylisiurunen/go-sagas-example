package message

import (
	"encoding/json"

	"github.com/markusylisiurunen/go-opinionatedevents"
)

type createShipmentTaskDataCustomer struct {
	Address string `json:"address"`
	Name    string `json:"name"`
}

type createShipmentTaskDataProduct struct {
	Price    int    `json:"price"`
	Quantity int    `json:"quantity"`
	UUID     string `json:"uuid"`
}

type createShipmentTaskData struct {
	Customer createShipmentTaskDataCustomer  `json:"customer"`
	Products []createShipmentTaskDataProduct `json:"products"`
}

type TaskCreateShipment struct {
	message
	Task createShipmentTaskData `json:"task"`
}

type TaskCreateShipmentParamsCustomer struct {
	Address string
	Name    string
}

type TaskCreateShipmentParamsProduct struct {
	Price    int
	Quantity int
	UUID     string
}

type TaskCreateShipmentParams struct {
	Customer TaskCreateShipmentParamsCustomer
	Products []TaskCreateShipmentParamsProduct
}

func NewTaskCreateShipment(args *TaskCreateShipmentParams, previous hasRollbackStack) *TaskCreateShipment {
	if previous == nil {
		previous = rollbackStack{}
	}
	products := []createShipmentTaskDataProduct{}
	for _, product := range args.Products {
		products = append(products, createShipmentTaskDataProduct(product))
	}
	return &TaskCreateShipment{
		message: message{
			Meta:          map[string]string{},
			RollbackStack: previous.getRollbackStack(),
		},
		Task: createShipmentTaskData{
			Customer: createShipmentTaskDataCustomer{
				Address: args.Customer.Address,
				Name:    args.Customer.Name,
			},
			Products: products,
		},
	}
}

func (m *TaskCreateShipment) Name() string {
	return "tasks.create_shipment"
}

func (m *TaskCreateShipment) MarshalPayload() ([]byte, error) {
	return json.Marshal(m)
}

func (m *TaskCreateShipment) UnmarshalPayload(data []byte) error {
	return json.Unmarshal(data, m)
}

func (m *TaskCreateShipment) GetOpinionated() *opinionatedevents.Message {
	v, err := opinionatedevents.NewMessage(m.Name(), m)
	if err != nil {
		panic(err)
	}
	return v
}
