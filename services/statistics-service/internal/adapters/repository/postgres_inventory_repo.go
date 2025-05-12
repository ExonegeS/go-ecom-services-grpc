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
	return runInTx(ctx, r.db, func(tx *sql.Tx) error {
		hourKey := fmt.Sprintf("%d", event.CreatedAt.Hour())
		_, err := tx.ExecContext(ctx, `
			INSERT INTO user_order_statistics (user_id, total_orders, hourly_distribution)
			VALUES ($1, 1, jsonb_build_object($2::text, 1))
			ON CONFLICT (user_id) DO UPDATE SET
			total_orders = user_order_statistics.total_orders + 1,
			hourly_distribution = user_order_statistics.hourly_distribution ||
				jsonb_build_object(
				$2::text,
				COALESCE((user_order_statistics.hourly_distribution->>$2)::INT, 0) + 1
				),
			updated_at = NOW()
		`, event.UserID, hourKey)
		if err != nil {
			return fmt.Errorf("order stats: %w", err)
		}

		var totalItems int64
		for _, it := range event.Items {
			totalItems += int64(it.Quantity)
		}
		var oldAOV float64
		var oldCount int64
		err = tx.QueryRowContext(ctx, `
			INSERT INTO user_statistics (user_id, total_items_purchased, total_completed_orders, average_order_value)
			VALUES ($1, $2, 1, $3)
			ON CONFLICT (user_id) DO UPDATE
			SET
				total_items_purchased = user_statistics.total_items_purchased + $2,
				total_completed_orders  = user_statistics.total_completed_orders + 1,
				average_order_value = (
				(user_statistics.average_order_value * user_statistics.total_completed_orders)
				+ $3
				) / (user_statistics.total_completed_orders + 1),
				updated_at = NOW()
			RETURNING average_order_value, total_completed_orders
		`, event.UserID, totalItems, event.Total).Scan(&oldAOV, &oldCount)
		if err != nil {
			return fmt.Errorf("user stats upsert: %w", err)
		}

		for _, it := range event.Items {
			_, err := tx.ExecContext(ctx, `
				INSERT INTO user_item_statistics (user_id, product_id, purchase_count)
				VALUES ($1, $2, $3)
				ON CONFLICT (user_id, product_id) DO UPDATE
				SET purchase_count = user_item_statistics.purchase_count + EXCLUDED.purchase_count
			`, event.UserID, it.ProductID, it.Quantity)
			if err != nil {
				return fmt.Errorf("item stats (%s): %w", it.ProductID, err)
			}
		}

		var topProduct string
		err = tx.QueryRowContext(ctx, `
			SELECT product_id
			FROM user_item_statistics
			WHERE user_id = $1
			ORDER BY purchase_count DESC
			LIMIT 1
		`, event.UserID).Scan(&topProduct)
		if err != nil && err != sql.ErrNoRows {
			return fmt.Errorf("select top item: %w", err)
		}

		_, err = tx.ExecContext(ctx, `
			UPDATE user_statistics
			SET most_purchased_item = $2,
				updated_at = NOW()
			WHERE user_id = $1
		`, event.UserID, topProduct)
		if err != nil {
			return fmt.Errorf("update most_purchased_item: %w", err)
		}

		return nil
	})
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
			COALESCE(most_purchased_item, '') AS most_purchased_item,
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
