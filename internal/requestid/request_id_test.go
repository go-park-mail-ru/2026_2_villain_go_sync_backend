package requestid

import (
	"context"
	"testing"
)

func TestFromContext(t *testing.T) {
	tests := []struct {
		name string
		ctx  context.Context
		want string
	}{
		{
			name: "id exists",
			ctx:  WithID(context.Background(), "abc123"),
			want: "abc123",
		},
		{
			name: "id missing",
			ctx:  context.Background(),
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FromContext(tt.ctx)
			if got != tt.want {
				t.Errorf("FromContext() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestWithID_Override(t *testing.T) {
	original := WithID(context.Background(), "first")
	updated := WithID(original, "second")

	if got := FromContext(original); got != "first" {
		t.Errorf("original context: got %q, want %q", got, "first")
	}

	if got := FromContext(updated); got != "second" {
		t.Errorf("updated context: got %q, want %q", got, "second")
	}
}
