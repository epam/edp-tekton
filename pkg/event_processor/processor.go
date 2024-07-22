package event_processor

import "context"

// Processor is an interface for event processors.
//
//go:generate mockery --name Processor --filename event_processor_mock.go
type Processor interface {
	Process(ctx context.Context, body []byte, ns, eventType string) (*EventInfo, error)
}
