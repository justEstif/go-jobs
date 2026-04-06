package domain

import "testing"

func TestNormalizeCompanyName(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"Google", "google"},
		{"Google, Inc.", "google"},
		{"Google LLC", "google"},
		{"Google Inc", "google"},
		{"  Google  ", "google"},
		{"JPMorgan Chase & Co.", "jpmorgan chase"},
		{"Stripe, Inc.", "stripe"},
		{"McKinsey & Company", "mckinsey"},
		{"Siemens AG", "siemens"},
		{"SAP S.A.", "sap"},
		{"", ""},
	}
	for _, tt := range tests {
		got := NormalizeCompanyName(tt.input)
		if got != tt.want {
			t.Errorf("NormalizeCompanyName(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestSlugify(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"Google", "google"},
		{"Stripe, Inc.", "stripe"},
		{"My Cool Startup", "my-cool-startup"},
		{"JPMorgan Chase & Co.", "jpmorgan-chase"},
		{"", ""},
	}
	for _, tt := range tests {
		got := Slugify(tt.input)
		if got != tt.want {
			t.Errorf("Slugify(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}
