package handler

import (
	"context"
	"time"

	"github.com/markusylisiurunen/go-opinionatedevents"
	"github.com/markusylisiurunen/sagas/cmd/message"
)

func handleTaskSendReceipt(
	ctx context.Context, publisher *opinionatedevents.Publisher, msg *message.TaskSendReceipt,
) opinionatedevents.ResultContainer {
	return handle(ctx, msg, func() error { return nil })
}

func HandleTaskSendReceipt(
	receiver *opinionatedevents.Receiver, publisher *opinionatedevents.Publisher,
) error {
	return receiver.On("default", "tasks.send_receipt",
		opinionatedevents.WithBackoff(opinionatedevents.LinearBackoff(2, 3, 15*time.Second))(
			message.WithRollback[message.TaskSendReceipt](publisher, 4)(
				func(ctx context.Context, _ string, delivery opinionatedevents.Delivery) opinionatedevents.ResultContainer {
					msg := &message.TaskSendReceipt{}
					if err := delivery.GetMessage().Payload(msg); err != nil {
						return opinionatedevents.FatalResult(err)
					}
					return handleTaskSendReceipt(ctx, publisher, msg)
				},
			),
		),
	)
}
