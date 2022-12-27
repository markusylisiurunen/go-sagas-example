package message

import (
	"encoding/json"

	"github.com/markusylisiurunen/go-opinionatedevents"
)

type reserveInventoryTaskDataCustomer struct {
	Address string `json:"address"`
	Name    string `json:"name"`
}

type reserveInventoryTaskDataProduct struct {
	Price    int    `json:"price"`
	Quantity int    `json:"quantity"`
	UUID     string `json:"uuid"`
}

type reserveInventoryTaskData struct {
	Customer reserveInventoryTaskDataCustomer  `json:"customer"`
	Products []reserveInventoryTaskDataProduct `json:"products"`
}

type TaskReserveInventory struct {
	message
	Task reserveInventoryTaskData `json:"task"`
}

type TaskReserveInventoryParamsCustomer struct {
	Address string
	Name    string
}

type TaskReserveInventoryParamsProduct struct {
	Price    int
	Quantity int
	UUID     string
}

type TaskReserveInventoryParams struct {
	Customer TaskReserveInventoryParamsCustomer
	Products []TaskReserveInventoryParamsProduct
}

func NewTaskReserveInventory(args *TaskReserveInventoryParams, previous hasRollbackStack) *TaskReserveInventory {
	if previous == nil {
		previous = rollbackStack{}
	}
	products := []reserveInventoryTaskDataProduct{}
	for _, product := range args.Products {
		products = append(products, reserveInventoryTaskDataProduct(product))
	}
	return &TaskReserveInventory{
		message: message{
			Meta:          map[string]string{},
			RollbackStack: previous.getRollbackStack(),
		},
		Task: reserveInventoryTaskData{
			Customer: reserveInventoryTaskDataCustomer{
				Address: args.Customer.Address,
				Name:    args.Customer.Name,
			},
			Products: products,
		},
	}
}

func (m *TaskReserveInventory) Name() string {
	return "tasks.reserve_inventory"
}

func (m *TaskReserveInventory) MarshalPayload() ([]byte, error) {
	return json.Marshal(m)
}

func (m *TaskReserveInventory) UnmarshalPayload(data []byte) error {
	return json.Unmarshal(data, m)
}

func (m *TaskReserveInventory) GetOpinionated() *opinionatedevents.Message {
	v, err := opinionatedevents.NewMessage(m.Name(), m)
	if err != nil {
		panic(err)
	}
	return v
}
