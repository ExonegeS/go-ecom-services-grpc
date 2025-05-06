package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/ExonegeS/go-ecom-services-grpc/services/statistics/internal/core/domain"
)

type PostgresStatisticsRepository struct {
	db *sql.DB
}

func NewPostgresStatisticsRepository(db *sql.DB) *PostgresStatisticsRepository {
	return &PostgresStatisticsRepository{db: db}
}

func (r *PostgresStatisticsRepository) UpdateOrderStatistics(ctx context.Context, event domain.OrderEvent) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Update user_order_statistics
	hour := event.CreatedAt.Hour()
	hourKey := fmt.Sprintf("%d", hour)

	_, err = tx.ExecContext(ctx, `
		INSERT INTO user_order_statistics (user_id, total_orders, hourly_distribution)
		VALUES ($1, 1, jsonb_build_object($2, 1))
		ON CONFLICT (user_id) DO UPDATE SET
			total_orders = user_order_statistics.total_orders + 1,
			hourly_distribution = user_order_statistics.hourly_distribution || jsonb_build_object($2, COALESCE(user_order_statistics.hourly_distribution->>$2::INT, 0) + 1),
			updated_at = NOW()
	`, event.UserID, hourKey)
	if err != nil {
		return fmt.Errorf("failed to update order stats: %w", err)
	}

	// Update user_statistics
	var totalItems int64
	for _, item := range event.Items {
		totalItems += int64(item.Quantity)
	}

	_, err = tx.ExecContext(ctx, `
		INSERT INTO user_statistics (user_id, total_items_purchased, total_completed_orders)
		VALUES ($1, $2, $3)
		ON CONFLICT (user_id) DO UPDATE SET
			total_items_purchased = user_statistics.total_items_purchased + $2,
			total_completed_orders = user_statistics.total_completed_orders + $3,
			updated_at = NOW()
	`, event.UserID, totalItems, 1)
	if err != nil {
		return fmt.Errorf("failed to update user stats: %w", err)
	}

	return tx.Commit()
}

func (r *PostgresStatisticsRepository) GetUserOrderStats(ctx context.Context, userID domain.UUID) (*domain.UserOrderStats, error) {
	var stats domain.UserOrderStats
	var hourlyJSON []byte

	err := r.db.QueryRowContext(ctx, `
		SELECT total_orders, hourly_distribution 
		FROM user_order_statistics 
		WHERE user_id = $1
	`, userID).Scan(&stats.TotalOrders, &hourlyJSON)
	if err != nil {
		if err == sql.ErrNoRows {
			return &domain.UserOrderStats{}, nil
		}
		return nil, fmt.Errorf("failed to get order stats: %w", err)
	}

	err = json.Unmarshal(hourlyJSON, &stats.HourlyDistribution)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal hourly distribution: %w", err)
	}

	return &stats, nil
}

func (r *PostgresStatisticsRepository) GetUserStatistics(ctx context.Context, userID domain.UUID) (*domain.UserStats, error) {
	var stats domain.UserStats

	err := r.db.QueryRowContext(ctx, `
		SELECT 
			total_items_purchased, 
			average_order_value, 
			most_purchased_item, 
			total_completed_orders 
		FROM user_statistics 
		WHERE user_id = $1
	`, userID).Scan(
		&stats.TotalItemsPurchased,
		&stats.AverageOrderValue,
		&stats.MostPurchasedItem,
		&stats.TotalCompletedOrders,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return &domain.UserStats{}, nil
		}
		return nil, fmt.Errorf("failed to get user stats: %w", err)
	}

	return &stats, nil
}
