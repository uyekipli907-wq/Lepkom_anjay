package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"html/template"
	"net/http"
	"strings"

	"golang.org/x/crypto/bcrypt"

	"Product_Inventory_437/middlewares"
	"Product_Inventory_437/models"
	"Product_Inventory_437/repositories"
)

type UserStore interface {
	FindByUsername(ctx context.Context, username string) (*models.User, error)
	FindByID(ctx context.Context, id int64) (*models.User, error)
	All(ctx context.Context) ([]models.User, error)
	Create(ctx context.Context, input models.UserInput) (*models.User, error)
	UpdateAPIToken(ctx context.Context, id int64, token string) error
}

type AuthHandler struct {
	users     UserStore
	auth      *middlewares.Authenticator
	templates *template.Template
}

func NewAuthHandler(users UserStore, auth *middlewares.Authenticator, templates *template.Template) *AuthHandler {
	return &AuthHandler{
		users:     users,
		auth:      auth,
		templates: templates,
	}
}

func (h *AuthHandler) ShowLogin(w http.ResponseWriter, r *http.Request) {
	data := map[string]any{
		"Title": "Login Product Inventory",
		"Error": r.URL.Query().Get("error"),
	}
	if err := h.templates.ExecuteTemplate(w, "login.html", data); err != nil {
		http.Error(w, "gagal merender halaman login", http.StatusInternalServerError)
	}
}

func (h *AuthHandler) LoginWeb(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		redirectWithMessage(w, r, "/login", "error", "form login tidak valid")
		return
	}

	user, err := h.authenticate(r.Context(), r.FormValue("username"), r.FormValue("password"))
	if err != nil {
		redirectWithMessage(w, r, "/login", "error", "username atau password salah")
		return
	}

	token, err := h.auth.GenerateToken(user)
	if err != nil {
		redirectWithMessage(w, r, "/login", "error", "gagal membuat sesi login")
		return
	}
	h.auth.SetAuthCookie(w, token)
	http.Redirect(w, r, "/products", http.StatusSeeOther)
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	h.auth.ClearAuthCookie(w)
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func (h *AuthHandler) LoginAPI(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeError(w, http.StatusBadRequest, "format json login tidak valid")
		return
	}

	user, err := h.authenticate(r.Context(), request.Username, request.Password)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "username atau password salah")
		return
	}

	token, err := h.auth.GenerateToken(user)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "gagal membuat token")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"message": "login berhasil",
		"token":   token,
		"user": map[string]any{
			"id":       user.ID,
			"username": user.Username,
			"role":     user.Role,
		},
	})
}

func (h *AuthHandler) RegisterAPI(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Role     string `json:"role"`
	}
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&request); err != nil {
		writeError(w, http.StatusBadRequest, "format json register tidak valid")
		return
	}

	username := strings.TrimSpace(request.Username)
	password := request.Password
	role := normalizeRole(request.Role)

	if username == "" {
		writeError(w, http.StatusBadRequest, "username wajib diisi")
		return
	}
	if len(password) < 6 {
		writeError(w, http.StatusBadRequest, "password minimal 6 karakter")
		return
	}
	if role == "" {
		writeError(w, http.StatusBadRequest, "role harus admin atau user")
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "gagal membuat password hash")
		return
	}

	user, err := h.users.Create(r.Context(), models.UserInput{
		Username:     username,
		PasswordHash: string(hash),
		Role:         role,
	})
	if errors.Is(err, repositories.ErrDuplicateUsername) {
		writeError(w, http.StatusConflict, "username sudah digunakan")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "gagal membuat user")
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{
		"message": "user berhasil dibuat",
		"data": map[string]any{
			"id":       user.ID,
			"username": user.Username,
			"role":     user.Role,
		},
	})
}

func (h *AuthHandler) UsersPage(w http.ResponseWriter, r *http.Request) {
	users, err := h.users.All(r.Context())
	if err != nil {
		http.Error(w, "gagal mengambil daftar user", http.StatusInternalServerError)
		return
	}

	currentUser, _ := middlewares.CurrentUser(r)
	data := map[string]any{
		"Title":   "User Management",
		"User":    currentUser,
		"Users":   users,
		"Success": r.URL.Query().Get("success"),
		"Error":   r.URL.Query().Get("error"),
	}
	if err := h.templates.ExecuteTemplate(w, "users.html", data); err != nil {
		http.Error(w, "gagal merender halaman user", http.StatusInternalServerError)
	}
}

func (h *AuthHandler) GenerateUserToken(w http.ResponseWriter, r *http.Request) {
	id, err := parsePathID(r, "id")
	if err != nil {
		redirectWithMessage(w, r, "/users", "error", err.Error())
		return
	}

	user, err := h.users.FindByID(r.Context(), id)
	if errors.Is(err, repositories.ErrNotFound) {
		redirectWithMessage(w, r, "/users", "error", "user tidak ditemukan")
		return
	}
	if err != nil {
		redirectWithMessage(w, r, "/users", "error", "gagal mengambil user")
		return
	}

	token, err := h.auth.GenerateToken(user)
	if err != nil {
		redirectWithMessage(w, r, "/users", "error", "gagal membuat token")
		return
	}

	if err := h.users.UpdateAPIToken(r.Context(), id, token); err != nil {
		redirectWithMessage(w, r, "/users", "error", "gagal menyimpan token")
		return
	}

	redirectWithMessage(w, r, "/users", "success", "token berhasil dibuat untuk "+user.Username)
}

func (h *AuthHandler) authenticate(ctx context.Context, username, password string) (*models.User, error) {
	username = strings.TrimSpace(username)
	if username == "" || password == "" {
		return nil, errors.New("credential kosong")
	}

	user, err := h.users.FindByUsername(ctx, username)
	if err != nil {
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, err
	}
	return user, nil
}

func normalizeRole(role string) string {
	switch strings.ToLower(strings.TrimSpace(role)) {
	case "", "user":
		return "user"
	case "admin":
		return "admin"
	default:
		return ""
	}
}
