package nats

import (
	"fmt"
	"log/slog"

	"github.com/ExonegeS/go-ecom-services-grpc/services/statistics/config"
	"github.com/ExonegeS/go-ecom-services-grpc/services/statistics/internal/core/domain"
	"github.com/ExonegeS/go-ecom-services-grpc/services/statistics/internal/core/ports"
	"github.com/nats-io/nats.go"
	"google.golang.org/protobuf/proto"

	orderspb "github.com/ExonegeS/go-ecom-services-grpc/services/statistics/internal/adapters/grpc"
)

type Subscriber struct {
	conn    *nats.Conn
	service ports.StatisticsService
	logger  *slog.Logger
}

func NewSubscriber(cfg config.NATSConfig, service ports.StatisticsService, logger *slog.Logger) (*Subscriber, error) {
	nc, err := nats.Connect(cfg.URL,
		nats.ErrorHandler(func(nc *nats.Conn, s *nats.Subscription, err error) {
			logger.Error("NATS error", "error", err)
		}),
		nats.DiscoveredServersHandler(func(nc *nats.Conn) {
			logger.Info("NATS discovered servers", "servers", nc.DiscoveredServers())
		}),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to NATS: %w", err)
	}

	logger.Info("Connected to NATS", "url", cfg.URL)
	return &Subscriber{
		conn:    nc,
		service: service,
		logger:  logger,
	}, nil
}

func (s *Subscriber) Subscribe() error {
	_, err := s.conn.Subscribe("orders.>", func(msg *nats.Msg) {
		s.logger.Debug("Received NATS message", "subject", msg.Subject)

		var event orderspb.OrderEvent
		if err := proto.Unmarshal(msg.Data, &event); err != nil {
			s.logger.Error("Failed to unmarshal order event", "error", err)
			return
		}

		domainEvent := convertProtoToDomainEvent(event)

		switch msg.Subject {
		case "orders.created":
			s.service.HandleOrderCreated(domainEvent)
		case "orders.updated":
			s.service.HandleOrderUpdated(domainEvent)
		case "orders.deleted":
			s.service.HandleOrderDeleted(domainEvent)
		default:
			s.logger.Warn("Received unknown order event", "subject", msg.Subject)
		}
	})
	return err
}

func convertProtoToDomainEvent(pbEvent orderspb.OrderEvent) domain.OrderEvent {
	return domain.OrderEvent{
		EventID:   pbEvent.GetEventId(),
		Operation: pbEvent.GetOperation(),
		OrderID:   pbEvent.GetOrderId(),
		UserID:    pbEvent.GetUserId(),
		Total:     pbEvent.GetTotal(),
		Status:    pbEvent.GetStatus(),
		CreatedAt: pbEvent.GetCreatedAt().AsTime(),
	}
}
