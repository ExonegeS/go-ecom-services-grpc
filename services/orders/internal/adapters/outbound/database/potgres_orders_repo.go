package database

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/ExonegeS/go-ecom-services-grpc/services/orders/internal/domain/entity"
	"github.com/ExonegeS/go-ecom-services-grpc/services/orders/internal/domain/ports"
)

type postgresOrdersRepository struct {
	db *sql.DB
}

// NewPostgresOrdersRepository returns a new OrdersRepository backed by Postgres.
func NewPostgresOrdersRepository(db *sql.DB) ports.OrdersRepository {
	return &postgresOrdersRepository{db: db}
}

func (r *postgresOrdersRepository) GetOrderByID(ctx context.Context, id entity.UUID) (*entity.Order, error) {
	const op = "postgresOrdersRepository.GetOrderByID"
	// fetch order
	var ord entity.Order
	err := r.db.QueryRowContext(ctx,
		`SELECT id, user_id, user_name, total_amount, status, created_at, updated_at
		 FROM orders WHERE id = $1`, id,
	).Scan(
		&ord.ID,
		&ord.UserID,
		&ord.UserName,
		&ord.TotalAmount,
		&ord.Status,
		&ord.CreatedAt,
		&ord.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, entity.ErrOrderNotFound
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	// fetch items
	rows, err := r.db.QueryContext(ctx,
		`SELECT product_id, product_name, product_price, quantity, created_at, updated_at
		 FROM order_items WHERE order_id = $1`, ord.ID,
	)
	if err != nil {
		return nil, fmt.Errorf("%s fetching items: %w", op, err)
	}
	defer rows.Close()

	ord.Items = make([]entity.OrderItem, 0)
	for rows.Next() {
		var it entity.OrderItem
		err = rows.Scan(
			&it.ProductID,
			&it.ProductName,
			&it.ProductPrice,
			&it.Quantity,
			&it.CreatedAt,
			&it.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("%s scanning item: %w", op, err)
		}
		ord.Items = append(ord.Items, it)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("%s items rows: %w", op, err)
	}

	return &ord, nil
}

func (r *postgresOrdersRepository) SaveOrder(ctx context.Context, ord entity.Order) error {
	const op = "postgresOrdersRepository.SaveOrder"
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("%s begin tx: %w", op, err)
	}
	// insert order
	_, err = tx.ExecContext(ctx,
		`INSERT INTO orders(id, user_id, user_name, total_amount, status, created_at, updated_at)
		 VALUES($1,$2,$3,$4,$5,$6,$7)`,
		ord.ID, ord.UserID, ord.UserName, ord.TotalAmount, ord.Status, ord.CreatedAt, ord.UpdatedAt,
	)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("%s insert order: %w", op, err)
	}
	// insert items
	for _, it := range ord.Items {
		_, err = tx.ExecContext(ctx,
			`INSERT INTO order_items(order_id, product_id, product_name, product_price, quantity, created_at, updated_at)
			 VALUES($1,$2,$3,$4,$5,$6,$7)`,
			ord.ID, it.ProductID, it.ProductName, it.ProductPrice, it.Quantity, it.CreatedAt, it.UpdatedAt,
		)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("%s insert item: %w", op, err)
		}
	}
	return tx.Commit()
}

func (r *postgresOrdersRepository) UpdateOrderByID(ctx context.Context, id entity.UUID, updateFn func(*entity.Order) (bool, error)) error {
	const op = "postgresOrdersRepository.UpdateOrderByID"
	// load existing
	ord, err := r.GetOrderByID(ctx, id)
	if err != nil {
		return err
	}
	// apply update function
	changed, err := updateFn(ord)
	if err != nil {
		return fmt.Errorf("%s updateFn: %w", op, err)
	}
	if !changed {
		return nil
	}
	// persist changes (only user_name, status, total_amount, updated_at)
	_, err = r.db.ExecContext(ctx,
		`UPDATE orders SET user_name=$1, status=$2, total_amount=$3, updated_at=$4 WHERE id=$5`,
		ord.UserName, ord.Status, ord.TotalAmount, ord.UpdatedAt, ord.ID,
	)
	if err != nil {
		return fmt.Errorf("%s exec update: %w", op, err)
	}
	return nil
}

func (r *postgresOrdersRepository) DeleteOrderByID(ctx context.Context, id entity.UUID) error {
	const op = "postgresOrdersRepository.DeleteOrderByID"
	// delete items, then order
	_, err := r.db.ExecContext(ctx, `DELETE FROM order_items WHERE order_id=$1`, id)
	if err != nil {
		return fmt.Errorf("%s delete items: %w", op, err)
	}
	_, err = r.db.ExecContext(ctx, `DELETE FROM orders WHERE id=$1`, id)
	if err != nil {
		return fmt.Errorf("%s delete order: %w", op, err)
	}
	return nil
}

func (r *postgresOrdersRepository) GetTotalOrdersCount(ctx context.Context) (int64, error) {
	const op = "postgresOrdersRepository.GetTotalOrdersCount"
	var cnt int64
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM orders`).Scan(&cnt)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}
	return cnt, nil
}

func (r *postgresOrdersRepository) GetAllOrders(ctx context.Context, pagination *entity.Pagination) ([]*entity.Order, error) {
	const op = "postgresOrdersRepository.GetAllOrders"
	// fetch paged orders
	query := `SELECT id, user_id, user_name, total_amount, status, created_at, updated_at
		 FROM orders ORDER BY created_at DESC LIMIT $1 OFFSET $2`
	rows, err := r.db.QueryContext(ctx, query, pagination.PageSize, pagination.Offset())
	if err != nil {
		return nil, fmt.Errorf("%s query orders: %w", op, err)
	}
	defer rows.Close()

	orders := make([]*entity.Order, 0)
	for rows.Next() {
		ord := new(entity.Order)
		err = rows.Scan(
			&ord.ID,
			&ord.UserID,
			&ord.UserName,
			&ord.TotalAmount,
			&ord.Status,
			&ord.CreatedAt,
			&ord.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("%s scan order: %w", op, err)
		}
		// optionally preload items
		orders = append(orders, ord)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("%s rows: %w", op, err)
	}
	return orders, nil
}
