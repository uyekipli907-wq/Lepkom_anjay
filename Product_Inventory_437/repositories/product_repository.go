package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"Product_Inventory_437/models"
)

type ProductRepository struct {
	db *sql.DB
}

func NewProductRepository(db *sql.DB) *ProductRepository {
	return &ProductRepository{db: db}
}

func (r *ProductRepository) All(ctx context.Context) ([]models.Product, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, name, COALESCE(description, ''), price, stock, created_at, updated_at
		FROM products
		ORDER BY id DESC`)
	if err != nil {
		return nil, fmt.Errorf("mengambil daftar produk: %w", err)
	}
	defer rows.Close()

	products := make([]models.Product, 0)
	for rows.Next() {
		var product models.Product
		if err := rows.Scan(
			&product.ID,
			&product.Name,
			&product.Description,
			&product.Price,
			&product.Stock,
			&product.CreatedAt,
			&product.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("membaca produk: %w", err)
		}
		products = append(products, product)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterasi produk: %w", err)
	}
	return products, nil
}

func (r *ProductRepository) ByID(ctx context.Context, id int64) (*models.Product, error) {
	var product models.Product
	err := r.db.QueryRowContext(ctx, `
		SELECT id, name, COALESCE(description, ''), price, stock, created_at, updated_at
		FROM products
		WHERE id = ?`, id).Scan(
		&product.ID,
		&product.Name,
		&product.Description,
		&product.Price,
		&product.Stock,
		&product.CreatedAt,
		&product.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("mengambil produk: %w", err)
	}
	return &product, nil
}

func (r *ProductRepository) Create(ctx context.Context, input models.ProductInput) (*models.Product, error) {
	result, err := r.db.ExecContext(ctx, `
		INSERT INTO products (name, description, price, stock)
		VALUES (?, ?, ?, ?)`,
		input.Name,
		input.Description,
		input.Price,
		input.Stock,
	)
	if err != nil {
		return nil, fmt.Errorf("menambahkan produk: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("membaca id produk baru: %w", err)
	}
	return r.ByID(ctx, id)
}

func (r *ProductRepository) Update(ctx context.Context, id int64, input models.ProductInput) (*models.Product, error) {
	result, err := r.db.ExecContext(ctx, `
		UPDATE products
		SET name = ?, description = ?, price = ?, stock = ?
		WHERE id = ?`,
		input.Name,
		input.Description,
		input.Price,
		input.Stock,
		id,
	)
	if err != nil {
		return nil, fmt.Errorf("mengubah produk: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("membaca jumlah update produk: %w", err)
	}
	if affected == 0 {
		return nil, ErrNotFound
	}
	return r.ByID(ctx, id)
}

func (r *ProductRepository) Delete(ctx context.Context, id int64) error {
	result, err := r.db.ExecContext(ctx, "DELETE FROM products WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("menghapus produk: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("membaca jumlah delete produk: %w", err)
	}
	if affected == 0 {
		return ErrNotFound
	}
	return nil
}
