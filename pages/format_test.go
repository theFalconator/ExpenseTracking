package pages_test

import (
	"testing"

	"github.com/theFalconator/expenses/pages"
)

func TestFormatCents(t *testing.T) {
	tests := []struct {
		name string
		cents int
		want  string
	}{
		{"Leading zero for dollars", 25, "$0.25"},
		{"Cents as dollars", 100, "$1.00"},
		{"larger value", 35012, "$350.12"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := pages.FormatCents(tt.cents)
			if got != tt.want {
				t.Errorf("FormatCents() = %v, want %v", got, tt.want)
			}
		})
	}
}


func TestFormatCentsNumber(t *testing.T) {
	tests := []struct {
		name string
		cents int
		want  string
	}{
		{"Leading zero for dollars", 25, "0.25"},
		{"Cents as dollars", 100, "1.00"},
		{"larger value", 35012, "350.12"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := pages.FormatCentsNumber(tt.cents)
			if got != tt.want {
				t.Errorf("FormatCentsNumber() = %v, want %v", got, tt.want)
			}
		})
	}
}

