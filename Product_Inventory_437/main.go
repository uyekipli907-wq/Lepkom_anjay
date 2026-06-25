package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"

	"Product_Inventory_437/configs"
	"Product_Inventory_437/handlers"
	"Product_Inventory_437/middlewares"
	"Product_Inventory_437/repositories"
	"Product_Inventory_437/routes"
)

func main() {
	cfg := configs.LoadConfig()

	db, err := configs.ConnectDatabase(cfg)
	if err != nil {
		log.Fatalf("database error: %v", err)
	}
	defer db.Close()

	templates, err := template.New("").Funcs(template.FuncMap{
		"rupiah": formatRupiah,
	}).ParseGlob("templates/*.html")
	if err != nil {
		log.Fatalf("template error: %v", err)
	}

	productRepo := repositories.NewProductRepository(db)
	userRepo := repositories.NewUserRepository(db)
	authenticator := middlewares.NewAuthenticator(cfg.JWTSecret, cfg.CookieName, userRepo)

	productHandler := handlers.NewProductHandler(productRepo, templates)
	authHandler := handlers.NewAuthHandler(userRepo, authenticator, templates)

	mux := http.NewServeMux()
	routes.Register(mux, routes.Dependencies{
		Products:      productHandler,
		Auth:          authHandler,
		Authenticator: authenticator,
	})

	handler := middlewares.Logger(mux)
	log.Printf("server berjalan di http://localhost:%s", cfg.AppPort)
	log.Fatal(http.ListenAndServe(cfg.ServerAddr(), handler))
}

func formatRupiah(value float64) string {
	return fmt.Sprintf("Rp %.2f", value)
}
