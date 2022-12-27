package message

import (
	"encoding/json"

	"github.com/markusylisiurunen/go-opinionatedevents"
)

type rollbackReserveInventoryTaskDataProduct struct {
	Quantity int    `json:"quantity"`
	UUID     string `json:"uuid"`
}

type rollbackReserveInventoryTaskData struct {
	Products []rollbackReserveInventoryTaskDataProduct `json:"products"`
}

type RollbackReserveInventory struct {
	message
	Task rollbackReserveInventoryTaskData `json:"task"`
}

type RollbackReserveInventoryParamsProduct struct {
	Quantity int
	UUID     string
}

type RollbackReserveInventoryParams struct {
	Products []RollbackReserveInventoryParamsProduct
}

func NewRollbackReserveInventory(args *RollbackReserveInventoryParams) *RollbackReserveInventory {
	products := []rollbackReserveInventoryTaskDataProduct{}
	for _, product := range args.Products {
		products = append(products, rollbackReserveInventoryTaskDataProduct(product))
	}
	return &RollbackReserveInventory{
		message: message{
			Meta:          map[string]string{},
			RollbackStack: rollbackStack{},
		},
		Task: rollbackReserveInventoryTaskData{
			Products: products,
		},
	}
}

func (m *RollbackReserveInventory) getRollbackStackItem() rollbackStackItem {
	return rollbackStackItem{
		Name: m.Name(),
		Meta: m.Meta,
		Data: map[string]any{
			"products": m.Task.Products,
		},
	}
}

func (m *RollbackReserveInventory) Name() string {
	return "tasks.rollback_reserve_inventory"
}

func (m *RollbackReserveInventory) MarshalPayload() ([]byte, error) {
	return json.Marshal(m)
}

func (m *RollbackReserveInventory) UnmarshalPayload(data []byte) error {
	return json.Unmarshal(data, m)
}

func (m *RollbackReserveInventory) GetOpinionated() *opinionatedevents.Message {
	v, err := opinionatedevents.NewMessage(m.Name(), m)
	if err != nil {
		panic(err)
	}
	return v
}
