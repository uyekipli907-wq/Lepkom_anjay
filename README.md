go mod init product_inventory

go get github.com/go-sql-driver/mysql
go get github.com/joho/godotenv
go get golang.org/x/crypto/bcrypt
go get github.com/golang-jwt/jwt/v5

go mod tidy

go run main.go
