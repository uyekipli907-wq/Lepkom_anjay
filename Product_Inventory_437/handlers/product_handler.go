package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"html/template"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"Product_Inventory_437/middlewares"
	"Product_Inventory_437/models"
	"Product_Inventory_437/repositories"
)

type ProductStore interface {
	All(ctx context.Context) ([]models.Product, error)
	ByID(ctx context.Context, id int64) (*models.Product, error)
	Create(ctx context.Context, input models.ProductInput) (*models.Product, error)
	Update(ctx context.Context, id int64, input models.ProductInput) (*models.Product, error)
	Delete(ctx context.Context, id int64) error
}

type ProductHandler struct {
	store     ProductStore
	templates *template.Template
}

func NewProductHandler(store ProductStore, templates *template.Template) *ProductHandler {
	return &ProductHandler{store: store, templates: templates}
}

func (h *ProductHandler) ListAPI(w http.ResponseWriter, r *http.Request) {
	products, err := h.store.All(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "gagal mengambil produk")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": products})
}

func (h *ProductHandler) CreateAPI(w http.ResponseWriter, r *http.Request) {
	input, err := decodeProductJSON(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	product, err := h.store.Create(r.Context(), input)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "gagal menambahkan produk")
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{
		"message": "produk berhasil ditambahkan",
		"data":    product,
	})
}

func (h *ProductHandler) UpdateAPI(w http.ResponseWriter, r *http.Request) {
	id, err := parsePathID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	input, err := decodeProductJSON(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	product, err := h.store.Update(r.Context(), id, input)
	if errors.Is(err, repositories.ErrNotFound) {
		writeError(w, http.StatusNotFound, "produk tidak ditemukan")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "gagal mengubah produk")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"message": "produk berhasil diubah",
		"data":    product,
	})
}

func (h *ProductHandler) DeleteAPI(w http.ResponseWriter, r *http.Request) {
	id, err := parsePathID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	err = h.store.Delete(r.Context(), id)
	if errors.Is(err, repositories.ErrNotFound) {
		writeError(w, http.StatusNotFound, "produk tidak ditemukan")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "gagal menghapus produk")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "produk berhasil dihapus"})
}

func (h *ProductHandler) WebList(w http.ResponseWriter, r *http.Request) {
	products, err := h.store.All(r.Context())
	if err != nil {
		http.Error(w, "gagal mengambil produk", http.StatusInternalServerError)
		return
	}

	user, _ := middlewares.CurrentUser(r)
	data := map[string]any{
		"Title":    "Product Inventory",
		"User":     user,
		"Products": products,
		"Success":  r.URL.Query().Get("success"),
		"Error":    r.URL.Query().Get("error"),
	}
	if err := h.templates.ExecuteTemplate(w, "products.html", data); err != nil {
		http.Error(w, "gagal merender halaman produk", http.StatusInternalServerError)
	}
}

func (h *ProductHandler) WebCreate(w http.ResponseWriter, r *http.Request) {
	input, err := parseProductForm(r)
	if err != nil {
		redirectWithMessage(w, r, "/products", "error", err.Error())
		return
	}

	if _, err := h.store.Create(r.Context(), input); err != nil {
		redirectWithMessage(w, r, "/products", "error", "gagal menambahkan produk")
		return
	}
	redirectWithMessage(w, r, "/products", "success", "produk berhasil ditambahkan")
}

func (h *ProductHandler) WebUpdate(w http.ResponseWriter, r *http.Request) {
	id, err := parsePathID(r, "id")
	if err != nil {
		redirectWithMessage(w, r, "/products", "error", err.Error())
		return
	}

	input, err := parseProductForm(r)
	if err != nil {
		redirectWithMessage(w, r, "/products", "error", err.Error())
		return
	}

	_, err = h.store.Update(r.Context(), id, input)
	if errors.Is(err, repositories.ErrNotFound) {
		redirectWithMessage(w, r, "/products", "error", "produk tidak ditemukan")
		return
	}
	if err != nil {
		redirectWithMessage(w, r, "/products", "error", "gagal mengubah produk")
		return
	}
	redirectWithMessage(w, r, "/products", "success", "produk berhasil diubah")
}

func (h *ProductHandler) WebDelete(w http.ResponseWriter, r *http.Request) {
	id, err := parsePathID(r, "id")
	if err != nil {
		redirectWithMessage(w, r, "/products", "error", err.Error())
		return
	}

	err = h.store.Delete(r.Context(), id)
	if errors.Is(err, repositories.ErrNotFound) {
		redirectWithMessage(w, r, "/products", "error", "produk tidak ditemukan")
		return
	}
	if err != nil {
		redirectWithMessage(w, r, "/products", "error", "gagal menghapus produk")
		return
	}
	redirectWithMessage(w, r, "/products", "success", "produk berhasil dihapus")
}

func decodeProductJSON(r *http.Request) (models.ProductInput, error) {
	var input models.ProductInput
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&input); err != nil {
		return input, errors.New("format json produk tidak valid")
	}
	return validateProductInput(input)
}

func parseProductForm(r *http.Request) (models.ProductInput, error) {
	if err := r.ParseForm(); err != nil {
		return models.ProductInput{}, errors.New("form produk tidak valid")
	}

	price, err := strconv.ParseFloat(r.FormValue("price"), 64)
	if err != nil {
		return models.ProductInput{}, errors.New("harga harus berupa angka")
	}

	stock, err := strconv.Atoi(r.FormValue("stock"))
	if err != nil {
		return models.ProductInput{}, errors.New("stok harus berupa angka bulat")
	}

	return validateProductInput(models.ProductInput{
		Name:        r.FormValue("name"),
		Description: r.FormValue("description"),
		Price:       price,
		Stock:       stock,
	})
}

func validateProductInput(input models.ProductInput) (models.ProductInput, error) {
	input.Name = strings.TrimSpace(input.Name)
	input.Description = strings.TrimSpace(input.Description)

	if input.Name == "" {
		return input, errors.New("nama produk wajib diisi")
	}
	if input.Price < 0 {
		return input, errors.New("harga produk tidak boleh negatif")
	}
	if input.Stock < 0 {
		return input, errors.New("stok produk tidak boleh negatif")
	}
	return input, nil
}

func parsePathID(r *http.Request, name string) (int64, error) {
	rawID := r.PathValue(name)
	id, err := strconv.ParseInt(rawID, 10, 64)
	if err != nil || id <= 0 {
		return 0, errors.New("id tidak valid")
	}
	return id, nil
}

func redirectWithMessage(w http.ResponseWriter, r *http.Request, path, key, message string) {
	values := url.Values{}
	values.Set(key, message)
	http.Redirect(w, r, path+"?"+values.Encode(), http.StatusSeeOther)
}
