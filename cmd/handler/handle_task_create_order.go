package handler

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/markusylisiurunen/go-opinionatedevents"
	"github.com/markusylisiurunen/sagas/cmd/message"
)

func handleTaskCreateOrder(
	ctx context.Context, publisher *opinionatedevents.Publisher, msg *message.TaskCreateOrder,
) opinionatedevents.ResultContainer {
	return handle(ctx, msg, func() error {
		// create the compensating message
		compensate := message.NewRollbackCreateOrder(&message.RollbackCreateOrderParams{
			Order: uuid.NewString(),
		})
		compensate.InjectFromMessage(msg)
		// create the next task
		products := []message.TaskReserveInventoryParamsProduct{}
		for _, product := range msg.Task.Products {
			products = append(products, message.TaskReserveInventoryParamsProduct(product))
		}
		task := message.NewTaskReserveInventory(&message.TaskReserveInventoryParams{
			Customer: message.TaskReserveInventoryParamsCustomer(msg.Task.Customer),
			Products: products,
		}, msg.RollbackStack.Push(compensate))
		task.InjectFromMessage(msg)
		// publish the message
		if err := publisher.Publish(ctx, task.GetOpinionated()); err != nil {
			return err
		}
		return nil
	})
}

func HandleTaskCreateOrder(
	receiver *opinionatedevents.Receiver, publisher *opinionatedevents.Publisher,
) error {
	return receiver.On("default", "tasks.create_order",
		opinionatedevents.WithBackoff(opinionatedevents.LinearBackoff(2, 3, 15*time.Second))(
			message.WithRollback[message.TaskCreateOrder](publisher, 4)(
				func(ctx context.Context, _ string, delivery opinionatedevents.Delivery) opinionatedevents.ResultContainer {
					msg := &message.TaskCreateOrder{}
					if err := delivery.GetMessage().Payload(msg); err != nil {
						return opinionatedevents.FatalResult(err)
					}
					return handleTaskCreateOrder(ctx, publisher, msg)
				},
			),
		),
	)
}
