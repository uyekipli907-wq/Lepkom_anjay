package configs

import (
	"database/sql"
	"errors"
	"fmt"
	"net"
	"regexp"
	"time"

	mysql "github.com/go-sql-driver/mysql"
	"golang.org/x/crypto/bcrypt"
)

var validDBName = regexp.MustCompile(`^[A-Za-z0-9_]+$`)

func ConnectDatabase(cfg Config) (*sql.DB, error) {
	if !validDBName.MatchString(cfg.DBName) {
		return nil, fmt.Errorf("nama database tidak valid: %s", cfg.DBName)
	}

	serverDB, err := sql.Open("mysql", mysqlDSN(cfg, ""))
	if err != nil {
		return nil, fmt.Errorf("membuka koneksi mysql: %w", err)
	}
	defer serverDB.Close()

	if err := serverDB.Ping(); err != nil {
		return nil, fmt.Errorf("mysql tidak dapat diakses: %w", err)
	}

	createDB := fmt.Sprintf("CREATE DATABASE IF NOT EXISTS `%s` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci", cfg.DBName)
	if _, err := serverDB.Exec(createDB); err != nil {
		return nil, fmt.Errorf("membuat database: %w", err)
	}

	db, err := sql.Open("mysql", mysqlDSN(cfg, cfg.DBName))
	if err != nil {
		return nil, fmt.Errorf("membuka koneksi database: %w", err)
	}

	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(30 * time.Minute)

	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("database tidak dapat diakses: %w", err)
	}

	if err := migrate(db); err != nil {
		_ = db.Close()
		return nil, err
	}

	if err := seed(db); err != nil {
		_ = db.Close()
		return nil, err
	}

	return db, nil
}

func mysqlDSN(cfg Config, dbName string) string {
	mysqlCfg := mysql.Config{
		User:                 cfg.DBUser,
		Passwd:               cfg.DBPassword,
		Net:                  "tcp",
		Addr:                 net.JoinHostPort(cfg.DBHost, cfg.DBPort),
		DBName:               dbName,
		ParseTime:            true,
		Loc:                  time.Local,
		AllowNativePasswords: true,
		Params: map[string]string{
			"charset": "utf8mb4",
		},
	}
	return mysqlCfg.FormatDSN()
}

func migrate(db *sql.DB) error {
	statements := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id INT AUTO_INCREMENT PRIMARY KEY,
			username VARCHAR(100) NOT NULL UNIQUE,
			password_hash VARCHAR(255) NOT NULL,
			role ENUM('admin', 'user') NOT NULL DEFAULT 'user',
			api_token TEXT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,
		`CREATE TABLE IF NOT EXISTS products (
			id INT AUTO_INCREMENT PRIMARY KEY,
			name VARCHAR(150) NOT NULL,
			description TEXT NULL,
			price DECIMAL(12,2) NOT NULL DEFAULT 0,
			stock INT NOT NULL DEFAULT 0,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,
		`CREATE TABLE IF NOT EXISTS stock_transaction (
			id INT AUTO_INCREMENT PRIMARY KEY,
			product_id INT NOT NULL,
			type ENUM('in', 'out') NOT NULL,
			quantity INT NOT NULL,
			note VARCHAR(255) NULL,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			CONSTRAINT fk_stock_transaction_product
				FOREIGN KEY (product_id) REFERENCES products(id)
				ON DELETE CASCADE
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,
	}

	for _, statement := range statements {
		if _, err := db.Exec(statement); err != nil {
			return fmt.Errorf("migrasi database gagal: %w", err)
		}
	}
	return nil
}

func seed(db *sql.DB) error {
	if err := seedUser(db, "admin", "password123", "admin"); err != nil {
		return err
	}
	if err := seedUser(db, "user1", "password123", "user"); err != nil {
		return err
	}
	return seedProducts(db)
}

func seedUser(db *sql.DB, username, password, role string) error {
	var id int64
	err := db.QueryRow("SELECT id FROM users WHERE username = ?", username).Scan(&id)
	if err == nil {
		return nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("memeriksa seed user %s: %w", username, err)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("membuat password hash: %w", err)
	}

	_, err = db.Exec(
		"INSERT INTO users (username, password_hash, role) VALUES (?, ?, ?)",
		username,
		string(hash),
		role,
	)
	if err != nil {
		return fmt.Errorf("menambahkan seed user %s: %w", username, err)
	}
	return nil
}

func seedProducts(db *sql.DB) error {
	var total int
	if err := db.QueryRow("SELECT COUNT(*) FROM products").Scan(&total); err != nil {
		return fmt.Errorf("menghitung seed produk: %w", err)
	}
	if total > 0 {
		return nil
	}

	products := []struct {
		name        string
		description string
		price       float64
		stock       int
	}{
		{"Keyboard Mechanical", "Keyboard kabel untuk operasional toko", 350000, 15},
		{"Mouse Wireless", "Mouse wireless ergonomis", 125000, 30},
		{"Monitor 24 Inch", "Monitor LED 24 inch full HD", 1450000, 8},
	}

	for _, product := range products {
		_, err := db.Exec(
			"INSERT INTO products (name, description, price, stock) VALUES (?, ?, ?, ?)",
			product.name,
			product.description,
			product.price,
			product.stock,
		)
		if err != nil {
			return fmt.Errorf("menambahkan seed produk %s: %w", product.name, err)
		}
	}
	return nil
}
