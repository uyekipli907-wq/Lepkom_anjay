package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	mysql "github.com/go-sql-driver/mysql"

	"Product_Inventory_437/models"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) FindByUsername(ctx context.Context, username string) (*models.User, error) {
	return r.findOne(ctx, `
		SELECT id, username, password_hash, role, COALESCE(api_token, ''), created_at, updated_at
		FROM users
		WHERE username = ?`, username)
}

func (r *UserRepository) FindByID(ctx context.Context, id int64) (*models.User, error) {
	return r.findOne(ctx, `
		SELECT id, username, password_hash, role, COALESCE(api_token, ''), created_at, updated_at
		FROM users
		WHERE id = ?`, id)
}

func (r *UserRepository) All(ctx context.Context) ([]models.User, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, username, password_hash, role, COALESCE(api_token, ''), created_at, updated_at
		FROM users
		ORDER BY id ASC`)
	if err != nil {
		return nil, fmt.Errorf("mengambil daftar user: %w", err)
	}
	defer rows.Close()

	users := make([]models.User, 0)
	for rows.Next() {
		var user models.User
		if err := rows.Scan(
			&user.ID,
			&user.Username,
			&user.PasswordHash,
			&user.Role,
			&user.APIToken,
			&user.CreatedAt,
			&user.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("membaca user: %w", err)
		}
		users = append(users, user)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterasi user: %w", err)
	}
	return users, nil
}

func (r *UserRepository) Create(ctx context.Context, input models.UserInput) (*models.User, error) {
	result, err := r.db.ExecContext(ctx, `
		INSERT INTO users (username, password_hash, role)
		VALUES (?, ?, ?)`,
		input.Username,
		input.PasswordHash,
		input.Role,
	)
	if err != nil {
		var mysqlErr *mysql.MySQLError
		if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
			return nil, ErrDuplicateUsername
		}
		return nil, fmt.Errorf("menambahkan user: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("membaca id user baru: %w", err)
	}
	return r.FindByID(ctx, id)
}

func (r *UserRepository) UpdateAPIToken(ctx context.Context, id int64, token string) error {
	result, err := r.db.ExecContext(ctx, "UPDATE users SET api_token = ? WHERE id = ?", token, id)
	if err != nil {
		return fmt.Errorf("menyimpan token api: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("membaca jumlah update token: %w", err)
	}
	if affected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *UserRepository) findOne(ctx context.Context, query string, args ...any) (*models.User, error) {
	var user models.User
	err := r.db.QueryRowContext(ctx, query, args...).Scan(
		&user.ID,
		&user.Username,
		&user.PasswordHash,
		&user.Role,
		&user.APIToken,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("mengambil user: %w", err)
	}
	return &user, nil
}
