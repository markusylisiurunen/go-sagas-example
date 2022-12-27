package handler

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/markusylisiurunen/go-opinionatedevents"
	"github.com/markusylisiurunen/sagas/cmd/message"
)

func handleTaskCreateShipment(
	ctx context.Context, publisher *opinionatedevents.Publisher, msg *message.TaskCreateShipment,
) opinionatedevents.ResultContainer {
	return handle(ctx, msg, func() error {
		// create the compensating message
		compensate := message.NewRollbackCreateShipment(&message.RollbackCreateShipmentParams{
			Shipment: uuid.NewString(),
		})
		compensate.InjectFromMessage(msg)
		// create the next task
		products := []message.TaskSendReceiptParamsProduct{}
		for _, product := range msg.Task.Products {
			products = append(products, message.TaskSendReceiptParamsProduct(product))
		}
		task := message.NewTaskSendReceipt(&message.TaskSendReceiptParams{
			Customer: message.TaskSendReceiptParamsCustomer(msg.Task.Customer),
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

func HandleTaskCreateShipment(
	receiver *opinionatedevents.Receiver, publisher *opinionatedevents.Publisher,
) error {
	return receiver.On("default", "tasks.create_shipment",
		opinionatedevents.WithBackoff(opinionatedevents.LinearBackoff(2, 3, 15*time.Second))(
			message.WithRollback[message.TaskCreateShipment](publisher, 4)(
				func(ctx context.Context, _ string, delivery opinionatedevents.Delivery) opinionatedevents.ResultContainer {
					msg := &message.TaskCreateShipment{}
					if err := delivery.GetMessage().Payload(msg); err != nil {
						return opinionatedevents.FatalResult(err)
					}
					return handleTaskCreateShipment(ctx, publisher, msg)
				},
			),
		),
	)
}
