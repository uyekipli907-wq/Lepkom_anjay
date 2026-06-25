package tests

import (
	"fmt"
	"html/template"
	"testing"
)

func TestTemplatesParse(t *testing.T) {
	_, err := template.New("").Funcs(template.FuncMap{
		"rupiah": func(value float64) string {
			return fmt.Sprintf("Rp %.2f", value)
		},
	}).ParseGlob("../templates/*.html")
	if err != nil {
		t.Fatalf("parse templates: %v", err)
	}
}