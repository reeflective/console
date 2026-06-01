package strutil

import (
	"strings"
	"testing"
)

func TestTemplate(t *testing.T) {
	tests := []struct {
		name string
		text string
		data any
		want string
	}{
		{
			name: "simple field",
			text: "Hello {{.Name}}",
			data: map[string]any{"Name": "world"},
			want: "Hello world",
		},
		{
			name: "trim func",
			text: "[{{trim .S}}]",
			data: map[string]any{"S": "  padded  "},
			want: "[padded]",
		},
		{
			name: "range over slice",
			text: "{{range .Items}}{{.}},{{end}}",
			data: map[string]any{"Items": []string{"a", "b", "c"}},
			want: "a,b,c,",
		},
		{
			name: "no substitution",
			text: "static text",
			data: nil,
			want: "static text",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var b strings.Builder
			if err := Template(&b, tc.text, tc.data); err != nil {
				t.Fatalf("Template(%q): unexpected error: %v", tc.text, err)
			}
			if got := b.String(); got != tc.want {
				t.Fatalf("Template(%q) = %q, want %q", tc.text, got, tc.want)
			}
		})
	}
}
