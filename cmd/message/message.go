package message

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/markusylisiurunen/go-opinionatedevents"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

// payloadable implements the interface that `opinionatedevents` requires
type Payloadable interface {
	Name() string
	MarshalPayload() ([]byte, error)
	UnmarshalPayload(data []byte) error
}

// Observable allows injecting and extracting telemetry information
type Observable interface {
	Inject(ctx context.Context)
	InjectFromMessage(msg Observable)
	Extract(ctx context.Context) context.Context
}

// opinionated allows the messages to be converted to an `opinionated` message
type opinionated interface {
	GetOpinionated() *opinionatedevents.Message
}

// rollbackable allows the message to either initialise or continue its rollback
type rollbackable interface {
	Rollback() (*anyRollbackMessage, bool)
}

// TaskMessage is a "forward" message in the Saga, and it contains the rollback stack
type TaskMessage interface {
	Payloadable
	Observable
	opinionated
	rollbackable
}

// RollbackMessage is a "backward" message in the Saga, and it contains the remaining rollback stack
type RollbackMessage interface {
	Payloadable
	Observable
	opinionated
	rollbackable
}

// message implements the Observable interface
type message struct {
	Meta          map[string]string `json:"meta"`
	RollbackStack rollbackStack     `json:"rollback_stack"`
}

func (m *message) Rollback() (*anyRollbackMessage, bool) {
	if m.RollbackStack == nil {
		return rollbackStack{}.Pop()
	}
	return m.RollbackStack.Pop()
}

func (m *message) Inject(ctx context.Context) {
	carrier := propagation.MapCarrier{}
	otel.GetTextMapPropagator().Inject(ctx, carrier)
	m.Meta = carrier
}

func (m *message) InjectFromMessage(msg Observable) {
	ctx := msg.Extract(context.Background())
	carrier := propagation.MapCarrier{}
	otel.GetTextMapPropagator().Inject(ctx, carrier)
	m.Meta = carrier
}

func (m *message) Extract(ctx context.Context) context.Context {
	if m.Meta == nil {
		return ctx
	}
	carrier := propagation.MapCarrier(m.Meta)
	return otel.GetTextMapPropagator().Extract(ctx, carrier)
}

// rollback stack
// ---

// hasRollbackStack is an interface for anything that contains a rollback stack
type hasRollbackStack interface {
	getRollbackStack() rollbackStack
}

// hasRollbackStackItem is an interface for anything that contains a single rollback stack item
type hasRollbackStackItem interface {
	getRollbackStackItem() rollbackStackItem
}

// rollbackStackItem is a single rollback message in the rollback stack
type rollbackStackItem struct {
	Name string            `json:"name"`
	Meta map[string]string `json:"meta"`
	Data map[string]any    `json:"data"`
}

// rollbackStack implement a stack of rollback messages
type rollbackStack []rollbackStackItem

// getRollbackStack implements the hasRollbackStack interface for the rollback stack itself
func (s rollbackStack) getRollbackStack() rollbackStack {
	return s
}

func (s rollbackStack) Push(i hasRollbackStackItem) rollbackStack {
	return append(rollbackStack{i.getRollbackStackItem()}, s...)
}

func (s rollbackStack) Pop() (*anyRollbackMessage, bool) {
	if len(s) == 0 {
		return nil, false
	}
	first, rest := s[0], s[1:]
	msg := &anyRollbackMessage{
		name:          first.Name,
		Meta:          first.Meta,
		Task:          first.Data,
		RollbackStack: rest,
	}
	return msg, true
}

// polymorphic rollback message
// ---

type anyRollbackMessage struct {
	name string `json:"-"`

	Meta          map[string]string `json:"meta"`
	Task          map[string]any    `json:"task"`
	RollbackStack rollbackStack     `json:"rollback_stack"`
}

func (m *anyRollbackMessage) MarshalPayload() ([]byte, error) {
	return json.Marshal(m)
}

func (m *anyRollbackMessage) UnmarshalPayload(data []byte) error {
	return errors.New("unexpected execution of UnmarshalPayload")
}

func (m *anyRollbackMessage) GetOpinionated() *opinionatedevents.Message {
	v, err := opinionatedevents.NewMessage(m.name, m)
	if err != nil {
		panic(err)
	}
	return v
}
