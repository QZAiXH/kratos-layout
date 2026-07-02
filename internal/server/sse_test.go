package server

import "testing"

func TestPathMatchesTemplate(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		template string
		want     bool
	}{
		{name: "exact", path: "/v1/todos/watch", template: "/v1/todos/watch", want: true},
		{name: "variable", path: "/v1/todos/42/events", template: "/v1/todos/{id}/events", want: true},
		{name: "empty variable", path: "/v1/todos//events", template: "/v1/todos/{id}/events", want: false},
		{name: "different", path: "/v1/todos/list", template: "/v1/todos/watch", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := pathMatchesTemplate(tt.path, tt.template); got != tt.want {
				t.Fatalf("pathMatchesTemplate(%q, %q) = %v, want %v", tt.path, tt.template, got, tt.want)
			}
		})
	}
}
