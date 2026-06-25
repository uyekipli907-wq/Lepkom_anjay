package tests

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"Product_Inventory_437/handlers"
	"Product_Inventory_437/models"
	"Product_Inventory_437/repositories"
)

type fakeProductStore struct {
	products []models.Product
	nextID   int64
}

func (f *fakeProductStore) All(ctx context.Context) ([]models.Product, error) {
	return f.products, nil
}

func (f *fakeProductStore) ByID(ctx context.Context, id int64) (*models.Product, error) {
	for _, product := range f.products {
		if product.ID == id {
			copy := product
			return &copy, nil
		}
	}
	return nil, repositories.ErrNotFound
}

func (f *fakeProductStore) Create(ctx context.Context, input models.ProductInput) (*models.Product, error) {
	if f.nextID == 0 {
		f.nextID = 1
	}

	product := models.Product{
		ID:          f.nextID,
		Name:        input.Name,
		Description: input.Description,
		Price:       input.Price,
		Stock:       input.Stock,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	f.nextID++
	f.products = append(f.products, product)
	return &product, nil
}

func (f *fakeProductStore) Update(ctx context.Context, id int64, input models.ProductInput) (*models.Product, error) {
	for i := range f.products {
		if f.products[i].ID == id {
			f.products[i].Name = input.Name
			f.products[i].Description = input.Description
			f.products[i].Price = input.Price
			f.products[i].Stock = input.Stock
			f.products[i].UpdatedAt = time.Now()
			return &f.products[i], nil
		}
	}
	return nil, repositories.ErrNotFound
}

func (f *fakeProductStore) Delete(ctx context.Context, id int64) error {
	for i := range f.products {
		if f.products[i].ID == id {
			f.products = append(f.products[:i], f.products[i+1:]...)
			return nil
		}
	}
	return repositories.ErrNotFound
}

func TestListProductsAPI(t *testing.T) {
	store := &fakeProductStore{
		products: []models.Product{
			{ID: 1, Name: "Keyboard", Description: "Keyboard USB", Price: 150000, Stock: 10},
			{ID: 2, Name: "Mouse", Description: "Mouse wireless", Price: 125000, Stock: 8},
		},
	}
	handler := handlers.NewProductHandler(store, nil)

	request := httptest.NewRequest(http.MethodGet, "/api/products", nil)
	response := httptest.NewRecorder()

	handler.ListAPI(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, response.Code)
	}

	var body struct {
		Data []models.Product `json:"data"`
	}
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(body.Data) != 2 {
		t.Fatalf("expected 2 products, got %d", len(body.Data))
	}
}

func TestCreateProductAPI(t *testing.T) {
	store := &fakeProductStore{}
	handler := handlers.NewProductHandler(store, nil)

	payload := `{"name":"Monitor","description":"Monitor 24 inch","price":1450000,"stock":5}`
	request := httptest.NewRequest(http.MethodPost, "/api/products", strings.NewReader(payload))
	response := httptest.NewRecorder()

	handler.CreateAPI(response, request)

	if response.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, response.Code)
	}
	if len(store.products) != 1 {
		t.Fatalf("expected 1 product created, got %d", len(store.products))
	}
	if store.products[0].Name != "Monitor" {
		t.Fatalf("expected product name Monitor, got %s", store.products[0].Name)
	}
}

func TestCreateProductAPIRejectsInvalidPayload(t *testing.T) {
	store := &fakeProductStore{}
	handler := handlers.NewProductHandler(store, nil)

	payload := `{"name":"","description":"invalid","price":-1,"stock":-2}`
	request := httptest.NewRequest(http.MethodPost, "/api/products", strings.NewReader(payload))
	response := httptest.NewRecorder()

	handler.CreateAPI(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, response.Code)
	}
	if len(store.products) != 0 {
		t.Fatalf("expected no product created, got %d", len(store.products))
	}
}
