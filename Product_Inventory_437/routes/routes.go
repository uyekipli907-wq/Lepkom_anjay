package routes

import (
	"net/http"

	"Product_Inventory_437/handlers"
	"Product_Inventory_437/middlewares"
)

type Dependencies struct {
	Products      *handlers.ProductHandler
	Auth          *handlers.AuthHandler
	Authenticator *middlewares.Authenticator
}

func Register(mux *http.ServeMux, deps Dependencies) {
	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	mux.HandleFunc("GET /login", deps.Auth.ShowLogin)
	mux.HandleFunc("POST /login", deps.Auth.LoginWeb)
	mux.Handle("POST /logout", deps.Authenticator.RequireAuth(http.HandlerFunc(deps.Auth.Logout)))
	mux.HandleFunc("POST /api/login", deps.Auth.LoginAPI)

	mux.Handle("GET /{$}", deps.Authenticator.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/products", http.StatusSeeOther)
	})))

	mux.Handle("GET /products", deps.Authenticator.RequireAuth(http.HandlerFunc(deps.Products.WebList)))
	mux.Handle("POST /products", deps.Authenticator.RequireAuth(http.HandlerFunc(deps.Products.WebCreate)))
	mux.Handle("POST /products/{id}/update", deps.Authenticator.RequireAuth(http.HandlerFunc(deps.Products.WebUpdate)))
	mux.Handle("POST /products/{id}/delete", deps.Authenticator.RequireAuth(http.HandlerFunc(deps.Products.WebDelete)))

	mux.Handle("GET /users", deps.Authenticator.RequireAdmin(http.HandlerFunc(deps.Auth.UsersPage)))
	mux.Handle("POST /users/{id}/token", deps.Authenticator.RequireAdmin(http.HandlerFunc(deps.Auth.GenerateUserToken)))

	mux.Handle("GET /api/products", deps.Authenticator.RequireAuth(http.HandlerFunc(deps.Products.ListAPI)))
	mux.Handle("POST /api/products", deps.Authenticator.RequireAuth(http.HandlerFunc(deps.Products.CreateAPI)))
	mux.Handle("PUT /api/products/{id}", deps.Authenticator.RequireAuth(http.HandlerFunc(deps.Products.UpdateAPI)))
	mux.Handle("DELETE /api/products/{id}", deps.Authenticator.RequireAuth(http.HandlerFunc(deps.Products.DeleteAPI)))
	mux.Handle("POST /api/register", deps.Authenticator.RequireAdmin(http.HandlerFunc(deps.Auth.RegisterAPI)))
}
