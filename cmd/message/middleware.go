package message

import (
	"context"
	"errors"
	"time"

	"github.com/markusylisiurunen/go-opinionatedevents"
)

func WithRollback[Message any, MessagePtr interface {
	TaskMessage
	*Message
}](
	publisher *opinionatedevents.Publisher, limit int,
) opinionatedevents.OnMessageMiddleware {
	doRollback := func(ctx context.Context, delivery opinionatedevents.Delivery) opinionatedevents.ResultContainer {
		// parse the message payload
		msg := MessagePtr(new(Message))
		if err := delivery.GetMessage().Payload(msg); err != nil {
			// TODO: what can be done here? there may be rollback messages in the stack but the payload cannot be parsed
			return opinionatedevents.FatalResult(err)
		}
		// publish the rollback message
		rollback, ok := msg.Rollback()
		if !ok {
			// the rollback stack is empty, processing is done
			err := errors.New("processing the task failed, nothing to roll back")
			return opinionatedevents.FatalResult(err)
		}
		if err := publisher.Publish(ctx, rollback.GetOpinionated()); err != nil {
			// processing this message will be retried until publishing the rollback message succeeds
			return opinionatedevents.ErrorResult(err, 2*time.Second)
		}
		// the rollback message was published successfully
		err := errors.New("processing the task failed, rolling back")
		return opinionatedevents.FatalResult(err)
	}

	return func(next opinionatedevents.OnMessageHandler) opinionatedevents.OnMessageHandler {
		return func(ctx context.Context, queue string, delivery opinionatedevents.Delivery) opinionatedevents.ResultContainer {
			if delivery.GetAttempt() > limit {
				// processing the message has already been attempted `limit` many times
				return doRollback(ctx, delivery)
			}
			// try to process the message
			result := next(ctx, queue, delivery)
			if result.GetResult().Err != nil && result.GetResult().RetryAt.IsZero() {
				// fatal error, should roll back immediately
				return doRollback(ctx, delivery)
			}
			if result.GetResult().Err != nil && delivery.GetAttempt() >= limit {
				// retriable error, but has failed too many times
				return doRollback(ctx, delivery)
			}
			// otherwise, don't roll back
			return result
		}
	}
}
