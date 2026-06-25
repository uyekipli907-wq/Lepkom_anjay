package middlewares

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"Product_Inventory_437/models"
)

type contextKey string

const authUserKey contextKey = "auth_user"

type UserLookup interface {
	FindByID(ctx context.Context, id int64) (*models.User, error)
}

type Claims struct {
	UserID   int64  `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

type Authenticator struct {
	secret     []byte
	cookieName string
	userLookup UserLookup
	tokenTTL   time.Duration
}

func NewAuthenticator(secret, cookieName string, userLookup UserLookup) *Authenticator {
	return &Authenticator{
		secret:     []byte(secret),
		cookieName: cookieName,
		userLookup: userLookup,
		tokenTTL:   24 * time.Hour,
	}
}

func (a *Authenticator) GenerateToken(user *models.User) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID:   user.ID,
		Username: user.Username,
		Role:     user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   strconv.FormatInt(user.ID, 10),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(a.tokenTTL)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(a.secret)
}

func (a *Authenticator) ParseToken(tokenString string) (*Claims, error) {
	claims := &Claims{}
	parser := jwt.NewParser(jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))

	token, err := parser.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (any, error) {
		return a.secret, nil
	})
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, errors.New("token tidak valid")
	}
	return claims, nil
}

func (a *Authenticator) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, err := a.userFromRequest(r)
		if err != nil {
			writeAuthError(w, r, http.StatusUnauthorized, "token tidak valid atau belum dikirim")
			return
		}

		ctx := context.WithValue(r.Context(), authUserKey, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (a *Authenticator) RequireAdmin(next http.Handler) http.Handler {
	return a.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, ok := CurrentUser(r)
		if !ok || user.Role != "admin" {
			writeAuthError(w, r, http.StatusForbidden, "akses ditolak: hanya admin yang diizinkan")
			return
		}
		next.ServeHTTP(w, r)
	}))
}

func (a *Authenticator) SetAuthCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     a.cookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int(a.tokenTTL.Seconds()),
	})
}

func (a *Authenticator) ClearAuthCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     a.cookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	})
}

func CurrentUser(r *http.Request) (*models.User, bool) {
	user, ok := r.Context().Value(authUserKey).(*models.User)
	return user, ok
}

func (a *Authenticator) userFromRequest(r *http.Request) (*models.User, error) {
	tokenString, ok := a.tokenFromRequest(r)
	if !ok {
		return nil, errors.New("token tidak ditemukan")
	}

	claims, err := a.ParseToken(tokenString)
	if err != nil {
		return nil, err
	}

	if a.userLookup == nil {
		return &models.User{
			ID:       claims.UserID,
			Username: claims.Username,
			Role:     claims.Role,
		}, nil
	}

	user, err := a.userLookup.FindByID(r.Context(), claims.UserID)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (a *Authenticator) tokenFromRequest(r *http.Request) (string, bool) {
	authHeader := r.Header.Get("Authorization")
	if strings.HasPrefix(strings.ToLower(authHeader), "bearer ") {
		token := strings.TrimSpace(authHeader[7:])
		return token, token != ""
	}

	cookie, err := r.Cookie(a.cookieName)
	if err == nil && cookie.Value != "" {
		return cookie.Value, true
	}

	return "", false
}

func writeAuthError(w http.ResponseWriter, r *http.Request, status int, message string) {
	if strings.HasPrefix(r.URL.Path, "/api/") {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": message})
		return
	}

	if status == http.StatusUnauthorized {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	http.Error(w, message, status)
}
