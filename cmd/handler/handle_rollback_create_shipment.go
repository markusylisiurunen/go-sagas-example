package handler

import (
	"context"
	"time"

	"github.com/markusylisiurunen/go-opinionatedevents"
	"github.com/markusylisiurunen/sagas/cmd/message"
)

func handleRollbackCreateShipment(
	ctx context.Context, publisher *opinionatedevents.Publisher, msg *message.RollbackCreateShipment,
) opinionatedevents.ResultContainer {
	return handle(ctx, msg, func() error {
		rollback, ok := msg.Rollback()
		if !ok {
			return nil
		}
		if err := publisher.Publish(ctx, rollback.GetOpinionated()); err != nil {
			return err
		}
		return nil
	})
}

func HandleRollbackCreateShipment(
	receiver *opinionatedevents.Receiver, publisher *opinionatedevents.Publisher,
) error {
	return receiver.On("default", "tasks.rollback_create_shipment",
		opinionatedevents.WithBackoff(opinionatedevents.LinearBackoff(2, 3, 15*time.Second))(
			func(ctx context.Context, _ string, delivery opinionatedevents.Delivery) opinionatedevents.ResultContainer {
				msg := &message.RollbackCreateShipment{}
				if err := delivery.GetMessage().Payload(msg); err != nil {
					return opinionatedevents.FatalResult(err)
				}
				return handleRollbackCreateShipment(ctx, publisher, msg)
			},
		),
	)
}
