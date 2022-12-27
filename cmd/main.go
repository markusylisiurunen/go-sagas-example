package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/markusylisiurunen/go-opinionatedevents"
	"github.com/markusylisiurunen/sagas/cmd/config"
	"github.com/markusylisiurunen/sagas/cmd/handler"
	"github.com/markusylisiurunen/sagas/cmd/message"
	"github.com/markusylisiurunen/sagas/internal/migrate"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
)

func initReceiver(ctx context.Context) (*opinionatedevents.Receiver, error) {
	receiver, err := opinionatedevents.NewReceiver()
	if err != nil {
		return nil, err
	}
	_, err = opinionatedevents.MakeReceiveFromPostgres(ctx, receiver, config.Default.DatabaseURL,
		opinionatedevents.ReceiveFromPostgresWithQueues("default"),
		opinionatedevents.ReceiveFromPostgresWithTableName("events"),
		opinionatedevents.ReceiveFromPostgresWithColumnNames(map[string]string{
			"deliver_at":        "event_deliver_at",
			"delivery_attempts": "event_delivery_attempts",
			"id":                "event_id",
			"name":              "event_name",
			"payload":           "event_data",
			"published_at":      "event_published_at",
			"queue":             "event_queue",
			"status":            "event_status",
			"topic":             "event_topic",
			"uuid":              "event_uuid",
		}),
		opinionatedevents.ReceiveFromPostgresWithIntervalTrigger(5*time.Second),
		opinionatedevents.ReceiveFromPostgresWithNotifyTrigger("__events"),
	)
	if err != nil {
		return nil, err
	}
	return receiver, nil
}

func initPublisher(_ context.Context) (*opinionatedevents.Publisher, error) {
	destination, err := opinionatedevents.NewPostgresDestination(config.Default.DatabaseURL,
		opinionatedevents.PostgresDestinationWithTopicToQueues("tasks", "default"),
		opinionatedevents.PostgresDestinationWithTableName("events"),
		opinionatedevents.PostgresDestinationWithColumnNames(map[string]string{
			"deliver_at":   "event_deliver_at",
			"id":           "event_id",
			"name":         "event_name",
			"payload":      "event_data",
			"published_at": "event_published_at",
			"queue":        "event_queue",
			"status":       "event_status",
			"topic":        "event_topic",
			"uuid":         "event_uuid",
		}),
	)
	if err != nil {
		return nil, err
	}
	publisher, err := opinionatedevents.NewPublisher(
		opinionatedevents.PublisherWithSyncBridge(destination),
	)
	if err != nil {
		return nil, err
	}
	return publisher, nil
}

func main() {
	rand.Seed(time.Now().UTC().UnixNano())
	ctx := context.Background()
	// run migrations
	if err := migrate.Migrate(ctx, "migrations", config.Default.DatabaseURL); err != nil {
		log.Fatal(err)
		return
	}
	// setup telemetry
	if err := setupTelemetry(ctx); err != nil {
		log.Fatal(err)
		return
	}
	// configure Pub/Sub
	receiver, err := initReceiver(ctx)
	if err != nil {
		log.Fatal(err)
		return
	}
	publisher, err := initPublisher(ctx)
	if err != nil {
		log.Fatal(err)
		return
	}
	// mount the handlers
	if err := handler.HandleTaskCreateOrder(receiver, publisher); err != nil {
		log.Fatal(err)
		return
	}
	if err := handler.HandleTaskReserveInventory(receiver, publisher); err != nil {
		log.Fatal(err)
		return
	}
	if err := handler.HandleTaskCreateShipment(receiver, publisher); err != nil {
		log.Fatal(err)
		return
	}
	if err := handler.HandleTaskSendReceipt(receiver, publisher); err != nil {
		log.Fatal(err)
		return
	}
	if err := handler.HandleRollbackCreateOrder(receiver, publisher); err != nil {
		log.Fatal(err)
		return
	}
	if err := handler.HandleRollbackReserveInventory(receiver, publisher); err != nil {
		log.Fatal(err)
		return
	}
	if err := handler.HandleRollbackCreateShipment(receiver, publisher); err != nil {
		log.Fatal(err)
		return
	}
	if err := handler.HandleRollbackSendReceipt(receiver, publisher); err != nil {
		log.Fatal(err)
		return
	}
	// periodically publish the saga's first task
	go func(ctx context.Context) {
		s, i := time.Duration(1), 0
		for {
			if i > 0 {
				s = time.Duration(15)
			}
			i += 1
			select {
			case <-ctx.Done():
				return
			case <-time.After(s * time.Second):
				// add one empty line for readability
				fmt.Println("")
				// start a trace for the saga
				ctx, span := otel.Tracer("sagas").Start(ctx, "sagas.process_order")
				// create the first task in the saga and inject the trace
				task := message.NewTaskCreateOrder(&message.TaskCreateOrderParams{
					Customer: message.TaskCreateOrderParamsCustomer{
						Address: "Aleksanterinkatu 52, 00100 Helsinki",
						Name:    "Essi Esimerkki",
					},
					Products: []message.TaskCreateOrderParamsProduct{
						{Price: 1299, Quantity: 1, UUID: uuid.NewString()},
						{Price: 4999, Quantity: 3, UUID: uuid.NewString()},
					},
				}, nil)
				task.Inject(ctx)
				// publish the saga
				if err := publisher.Publish(ctx, task.GetOpinionated()); err != nil {
					span.SetStatus(codes.Error, "error publishing the saga")
					span.RecordError(err)
					span.End()
					continue
				}
				span.End()
			}
		}
	}(ctx)
	// wait for the kill signal
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig
}
