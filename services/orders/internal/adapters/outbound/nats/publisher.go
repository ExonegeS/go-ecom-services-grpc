package nats

import (
	"context"
	"fmt"
	"time"

	orderspb "github.com/ExonegeS/go-ecom-services-grpc/services/orders/internal/adapters/outbound/grpc/statistics"
	"github.com/nats-io/nats.go"
	"google.golang.org/protobuf/proto"
)

type OrderEventPublisher struct {
	conn    *nats.Conn
	timeout time.Duration
}

func NewOrderEventPublisher(natsURL string) (*OrderEventPublisher, error) {
	nc, err := nats.Connect(natsURL, nats.ErrorHandler(func(nc *nats.Conn, s *nats.Subscription, err error) {
		fmt.Printf("NATS Error: %v", err)
	}))
	if err != nil {
		return nil, fmt.Errorf("NATS connection failed:", err)
	}
	fmt.Println("Connected to NATS at:", nc.ConnectedUrl())

	return &OrderEventPublisher{
		conn:    nc,
		timeout: 2 * time.Second,
	}, nil
}

func (p *OrderEventPublisher) PublishOrderCreated(ctx context.Context, order *orderspb.OrderEvent) error {
	return p.publishEvent(ctx, "orders.created", order)
}

func (p *OrderEventPublisher) PublishOrderUpdated(ctx context.Context, order *orderspb.OrderEvent) error {
	return p.publishEvent(ctx, "orders.updated", order)
}

func (p *OrderEventPublisher) PublishOrderDeleted(ctx context.Context, order *orderspb.OrderEvent) error {
	return p.publishEvent(ctx, "orders.deleted", order)
}

func (p *OrderEventPublisher) publishEvent(ctx context.Context, subject string, event proto.Message) error {
	data, err := proto.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	err = p.conn.Publish(subject, data)
	if err != nil {
		fmt.Printf("Failed to publish to %s: %v", subject, err)
	}
	return err
}

func (p *OrderEventPublisher) Close() {
	p.conn.Close()
}
