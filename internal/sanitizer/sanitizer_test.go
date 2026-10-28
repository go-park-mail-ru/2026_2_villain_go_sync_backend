package sanitizer

import (
	"encoding/json"
	"testing"
)

func TestSafeString_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want string
	}{
		{"plain text", `"Hello"`, "Hello"},
		{"script tag", `"<script>alert(1)</script>Hello"`, "Hello"},
		{"img onerror", `"<img src=x onerror=alert(1)>"`, ""},
		{"bold tags", `"<b>bold</b>"`, "bold"},
		{"cyrillic", `"Привет"`, "Привет"},
		{"empty", `""`, ""},
		{"JS command", `"<img src=x onerror=alert(1)>alert"`, "alert"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var s SafeString
			if err := json.Unmarshal([]byte(tt.raw), &s); err != nil {
				t.Fatalf("unmarshal: %v", err)
			}
			if s.String() != tt.want {
				t.Errorf("got %q, want %q", s.String(), tt.want)
			}
		})
	}
}
