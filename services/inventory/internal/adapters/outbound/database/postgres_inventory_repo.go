package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/ExonegeS/go-ecom-services-grpc/services/inventory/internal/adapters/outbound/database/model"
	"github.com/ExonegeS/go-ecom-services-grpc/services/inventory/internal/domain/entity"
	"github.com/ExonegeS/go-ecom-services-grpc/services/inventory/internal/domain/ports"
)

type postgresInventoryRepository struct {
	db *sql.DB
}

func NewPostgresInventoryRepository(db *sql.DB) ports.InventoryRepository {
	return &postgresInventoryRepository{db: db}
}

func (r *postgresInventoryRepository) GetByID(ctx context.Context, id entity.UUID) (*entity.InventoryItem, error) {
	const op = "postgresInventoryRepository.GetByID"

	var product model.Product
	var category model.Category
	err := r.db.QueryRowContext(ctx,
		`SELECT 
			id, 
			name, 
			description, 
			category_id, 
			price, 
			stock_quantity, 
			unit, 
			created_at, 
			updated_at 
		FROM products 
		WHERE id = $1`, id,
	).Scan(
		&product.ID,
		&product.Name,
		&product.Description,
		&product.CategoryID,
		&product.Price,
		&product.Stock,
		&product.Unit,
		&product.CreatedAt,
		&product.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, model.ErrProductNotFound
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	if product.CategoryID.Valid {
		err = r.db.QueryRowContext(ctx,
			`SELECT 
				id, 
				name, 
				description, 
				created_at, 
				updated_at 
			FROM categories 
			WHERE id = $1`, product.CategoryID,
		).Scan(
			&category.ID,
			&category.Name,
			&category.Description,
			&category.CreatedAt,
			&category.UpdatedAt,
		)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, model.ErrCategoryNotFound
			}
			return nil, fmt.Errorf("%s: %w", op, err)
		}
	}
	e, err := model.ModelToInventoryItem(&product, &category)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return e, nil
}

func (r *postgresInventoryRepository) Save(ctx context.Context, item entity.InventoryItem) error {
	const op = "postgresInventoryRepository.Save"

	return runInTx(ctx, r.db, func(tx *sql.Tx) error {
		m, _, err := model.InventoryItemToModel(&item)
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
		_, err = tx.ExecContext(ctx,
			`INSERT INTO products (id, name, description, category_id, price, stock_quantity, unit, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
			m.ID, m.Name, m.Description,
			m.CategoryID, m.Price, m.Stock,
			m.Unit, m.CreatedAt, m.UpdatedAt,
		)
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
		return nil
	})
}

func (r *postgresInventoryRepository) UpdateByID(ctx context.Context, id entity.UUID, updateFn func(*entity.InventoryItem) (bool, error)) error {
	const op = "postgresInventoryRepository.UpdateByID"
	return runInTx(ctx, r.db, func(tx *sql.Tx) error {
		var product model.Product
		var category model.Category
		err := r.db.QueryRowContext(ctx,
			`SELECT 
				name, 
				description, 
				category_id,
				price, 
				stock_quantity, 
				unit, 
				created_at 
			FROM products 
			WHERE id = $1 FOR UPDATE`, id,
		).Scan(
			&product.Name,
			&product.Description,
			&product.CategoryID,
			&product.Price,
			&product.Stock,
			&product.Unit,
			&product.CreatedAt,
		)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return model.ErrProductNotFound
			}
			return fmt.Errorf("%s: %w", op, err)
		}

		if product.CategoryID.Valid {
			err = r.db.QueryRowContext(ctx,
				`SELECT id, name, description FROM categories WHERE id = $1`, product.CategoryID,
			).Scan(
				&category.ID,
				&category.Name,
				&category.Description,
			)
			if err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					return model.ErrCategoryNotFound
				}
				return fmt.Errorf("%s: %w", op, err)
			}
		}

		e, err := model.ModelToInventoryItem(&product, &category)
		if err != nil {
			return err
		}

		updated, err := updateFn(e)
		if err != nil {
			return err
		}

		if !updated {
			return entity.ErrNotUpdated
		}

		p, _, err := model.InventoryItemToModel(e)
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}

		_, err = tx.ExecContext(ctx,
			"UPDATE products SET name = $1, description = $2, category_id = $3, price = $4, stock_quantity = $5, unit = $6, updated_at = $7 WHERE id = $8",
			p.Name, p.Description, p.CategoryID, p.Price, p.Stock, p.Unit, p.UpdatedAt, id)
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}

		return nil
	})
}

func (r *postgresInventoryRepository) DeleteByID(ctx context.Context, id entity.UUID) error {
	const op = "postgresInventoryRepository.DeleteByID"

	return runInTx(ctx, r.db, func(tx *sql.Tx) error {
		_, err := tx.ExecContext(ctx,
			`DELETE FROM products WHERE id = $1`, id,
		)
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
		return nil
	})
}

func (r *postgresInventoryRepository) GetAllInventoryItems(ctx context.Context, pagination *entity.Pagination) ([]*entity.InventoryItem, error) {
	const op = "postgresInventoryRepository.GetAllInventoryItems"

	query := `
		SELECT id, name, description, category_id, price, stock_quantity, unit, created_at, updated_at
		FROM products
	`

	if pagination.SortBy != "" {
		if pagination.SortBy == entity.SortByQuantity {
			pagination.SortBy = "stock_quantity"
		}
		query += fmt.Sprintf(" ORDER BY %s", pagination.SortBy)
	}

	offset := (pagination.Page - 1) * pagination.PageSize
	query += fmt.Sprintf(" LIMIT %d OFFSET %d", pagination.PageSize, offset)

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	var modelItems []model.Product
	for rows.Next() {
		var model model.Product
		err := rows.Scan(
			&model.ID,
			&model.Name,
			&model.Description,
			&model.CategoryID,
			&model.Price,
			&model.Stock,
			&model.Unit,
			&model.CreatedAt,
			&model.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		modelItems = append(modelItems, model)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	entities := make([]*entity.InventoryItem, 0, len(modelItems))
	for _, m := range modelItems {
		c, err := r.fetchCategoryData(ctx, entity.UUID(m.CategoryID.String))
		cPtr := &c
		if err != nil {
			cPtr = nil
		}
		e, err := model.ModelToInventoryItem(&m, cPtr)
		if err == nil {
			entities = append(entities, e)
		}
	}

	return entities, nil
}

func (r *postgresInventoryRepository) fetchCategoryData(ctx context.Context, categoryID entity.UUID) (category model.Category, err error) {
	const op = "postgresInventoryRepository.fetchCategoryData"

	err = r.db.QueryRowContext(ctx,
		`SELECT id, name, description, created_at, updated_at FROM categories WHERE id = $1`, categoryID,
	).Scan(
		&category.ID,
		&category.Name,
		&category.Description,
		&category.CreatedAt,
		&category.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = model.ErrCategoryNotFound
			return
		}
		err = fmt.Errorf("%s: %w", op, err)
	}
	return
}

func (r *postgresInventoryRepository) GetTotalCount(ctx context.Context) (int64, error) {
	const op = "postgresInventoryRepository.GetTotalCount"

	var total int64
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM products").Scan(&total)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}
	return total, nil
}

func (r *postgresInventoryRepository) GetCategoryByID(ctx context.Context, id entity.UUID) (*entity.Category, error) {
	const op = "postgresInventoryRepository.GetCategoryByID"

	var category model.Category

	err := r.db.QueryRowContext(ctx,
		`SELECT id, name, description, created_at, updated_at FROM categories WHERE id = $1`, id,
	).Scan(
		&category.ID,
		&category.Name,
		&category.Description,
		&category.CreatedAt,
		&category.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, model.ErrCategoryNotFound
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	e, err := model.ModelToCategory(&category)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return e, nil
}

func (r *postgresInventoryRepository) SaveCategory(ctx context.Context, category entity.Category) error {
	op := "postgresInventoryRepository.SaveCategory"

	return runInTx(ctx, r.db, func(tx *sql.Tx) error {
		m := model.CategoryToModel(&category)

		_, err := tx.ExecContext(ctx,
			`INSERT INTO categories (id, name, description, created_at, updated_at)
            VALUES ($1, $2, $3, $4, $5)`,
			m.ID, m.Name, m.Description,
			m.CreatedAt, m.UpdatedAt,
		)
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
		return nil
	})
}

func (r *postgresInventoryRepository) UpdateCategoryByID(ctx context.Context, id entity.UUID, updateFn func(*entity.Category) (bool, error)) error {
	const op = "postgresInventoryRepository.UpdateCategoryByID"
	return runInTx(ctx, r.db, func(tx *sql.Tx) error {
		var category model.Category
		err := r.db.QueryRowContext(ctx,
			`SELECT name, description, created_at FROM categories WHERE id = $1 FOR UPDATE`, id,
		).Scan(
			&category.Name,
			&category.Description,
			&category.CreatedAt,
		)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return model.ErrCategoryNotFound
			}
			return fmt.Errorf("%s: %w", op, err)
		}

		e, err := model.ModelToCategory(&category)
		if err != nil {
			return err
		}

		updated, err := updateFn(e)
		if err != nil {
			return err
		}

		if !updated {
			return entity.ErrNotUpdated
		}

		c := model.CategoryToModel(e)

		_, err = tx.ExecContext(ctx,
			"UPDATE categories SET name = $1, description = $2, updated_at = $3 WHERE id = $4",
			c.Name, c.Description, c.UpdatedAt, id)
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}

		return nil
	})
}

func (r *postgresInventoryRepository) DeleteCategoryByID(ctx context.Context, id entity.UUID) error {
	const op = "postgresInventoryRepository.DeleteCategoryByID"

	return runInTx(ctx, r.db, func(tx *sql.Tx) error {
		_, err := tx.ExecContext(ctx,
			`DELETE FROM categories WHERE id = $1`, id,
		)
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
		return nil
	})
}

func (r *postgresInventoryRepository) GetTotalCategoriesCount(ctx context.Context) (int64, error) {
	const op = "postgresInventoryRepository.GetTotalCategoriesCount"

	var total int64
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM categories").Scan(&total)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}
	return total, nil
}

func (r *postgresInventoryRepository) GetAllCategories(ctx context.Context, pagination *entity.Pagination) ([]*entity.Category, error) {
	const op = "postgresInventoryRepository.GetAllInventoryItems"

	query := `
		SELECT id, name, description, created_at, updated_at
		FROM categories
	`

	if pagination.SortBy != "" {
		query += fmt.Sprintf(" ORDER BY %s", pagination.SortBy)
	}

	offset := (pagination.Page - 1) * pagination.PageSize
	query += fmt.Sprintf(" LIMIT %d OFFSET %d", pagination.PageSize, offset)

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	var modelItems []model.Category
	for rows.Next() {
		var model model.Category
		err := rows.Scan(
			&model.ID,
			&model.Name,
			&model.Description,
			&model.CreatedAt,
			&model.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		modelItems = append(modelItems, model)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	entities := make([]*entity.Category, 0, len(modelItems))
	for _, m := range modelItems {
		e, err := model.ModelToCategory(&m)
		if err == nil {
			entities = append(entities, e)
		}
	}

	return entities, nil
}
