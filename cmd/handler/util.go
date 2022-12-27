package handler

import (
	"context"
	"errors"
	"log"
	"math/rand"
	"time"

	"github.com/markusylisiurunen/go-opinionatedevents"
	"github.com/markusylisiurunen/sagas/cmd/message"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
)

func handle[Message any, MessagePtr interface {
	*Message
	message.Payloadable
	message.Observable
}](ctx context.Context, msg MessagePtr, next func() error) opinionatedevents.ResultContainer {
	_, span := otel.Tracer("sagas").Start(msg.Extract(ctx), msg.Name())
	defer span.End()

	log.Printf("%s: begin", msg.Name())
	defer log.Printf("%s: end", msg.Name())

	time.Sleep(time.Duration(rand.Intn(300)+200) * time.Millisecond)

	if rand.Float64() < 0.5 {
		err := errors.New("something went wrong")
		log.Printf("%s: %s", msg.Name(), err.Error())
		span.SetStatus(codes.Error, err.Error())
		return opinionatedevents.ErrorResult(err, 10*time.Second)
	}

	if err := next(); err != nil {
		log.Printf("%s: %s", msg.Name(), err.Error())
		span.SetStatus(codes.Error, err.Error())
		return opinionatedevents.ErrorResult(err, 10*time.Second)
	}

	return opinionatedevents.SuccessResult()
}
