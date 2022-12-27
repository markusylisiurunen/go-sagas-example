package handler

import (
	"context"
	"time"

	"github.com/markusylisiurunen/go-opinionatedevents"
	"github.com/markusylisiurunen/sagas/cmd/message"
)

func handleTaskReserveInventory(
	ctx context.Context, publisher *opinionatedevents.Publisher, msg *message.TaskReserveInventory,
) opinionatedevents.ResultContainer {
	return handle(ctx, msg, func() error {
		// create the compensating message
		productsForRollback := []message.RollbackReserveInventoryParamsProduct{}
		for _, product := range msg.Task.Products {
			productsForRollback = append(productsForRollback, message.RollbackReserveInventoryParamsProduct{
				Quantity: product.Quantity,
				UUID:     product.UUID,
			})
		}
		compensate := message.NewRollbackReserveInventory(&message.RollbackReserveInventoryParams{
			Products: productsForRollback,
		})
		compensate.InjectFromMessage(msg)
		// create the next task
		productsForTask := []message.TaskCreateShipmentParamsProduct{}
		for _, product := range msg.Task.Products {
			productsForTask = append(productsForTask, message.TaskCreateShipmentParamsProduct(product))
		}
		task := message.NewTaskCreateShipment(&message.TaskCreateShipmentParams{
			Customer: message.TaskCreateShipmentParamsCustomer(msg.Task.Customer),
			Products: productsForTask,
		}, msg.RollbackStack.Push(compensate))
		task.InjectFromMessage(msg)
		// publish the message
		if err := publisher.Publish(ctx, task.GetOpinionated()); err != nil {
			return err
		}
		return nil
	})
}

func HandleTaskReserveInventory(
	receiver *opinionatedevents.Receiver, publisher *opinionatedevents.Publisher,
) error {
	return receiver.On("default", "tasks.reserve_inventory",
		opinionatedevents.WithBackoff(opinionatedevents.LinearBackoff(2, 3, 15*time.Second))(
			message.WithRollback[message.TaskReserveInventory](publisher, 4)(
				func(ctx context.Context, _ string, delivery opinionatedevents.Delivery) opinionatedevents.ResultContainer {
					msg := &message.TaskReserveInventory{}
					if err := delivery.GetMessage().Payload(msg); err != nil {
						return opinionatedevents.FatalResult(err)
					}
					return handleTaskReserveInventory(ctx, publisher, msg)
				},
			),
		),
	)
}
