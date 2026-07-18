package line

import (
	"errors"
	"reflect"
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    []string
		wantErr bool
	}{
		{"empty", "", nil, false},
		{"simple", "echo hello", []string{"echo", "hello"}, false},
		{"extra spaces collapse", "echo   hello    world", []string{"echo", "hello", "world"}, false},
		{"trailing comment", "echo hello # a comment", []string{"echo", "hello"}, false},
		{"comment only", "# just a comment", nil, false},
		{"single quotes", "echo 'hello world'", []string{"echo", "hello world"}, false},
		{"double quotes", `echo "hello world"`, []string{"echo", "hello world"}, false},
		{"quoted hash not a comment", "echo '# not a comment'", []string{"echo", "# not a comment"}, false},
		{"unterminated single quote", "echo 'oops", nil, true},
		{"unterminated double quote", `echo "oops`, nil, true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := Parse(tc.input, EscapeShell)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("Parse(%q): expected error, got nil (words=%q)", tc.input, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("Parse(%q): unexpected error: %v", tc.input, err)
			}
			if len(got) == 0 && len(tc.want) == 0 {
				return
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("Parse(%q) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}

func TestSplit(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantWords []string
		wantErr   error
	}{
		{"empty", "", []string{}, nil},
		{"simple", "echo hello", []string{"echo", "hello"}, nil},
		{"single quotes", "echo 'hello world'", []string{"echo", "hello world"}, nil},
		{"double quotes", `echo "hello world"`, []string{"echo", "hello world"}, nil},
		{"escaped space", `echo foo\ bar`, []string{"echo", "foo bar"}, nil},
		{"unterminated single", "echo 'oops", []string{"echo"}, ErrUnterminatedSingleQuote},
		{"unterminated double", `echo "oops`, []string{"echo"}, ErrUnterminatedDoubleQuote},
		{"trailing backslash", `echo foo\`, []string{"echo"}, ErrUnterminatedEscape},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			words, _, err := Split(tc.input, false, EscapeShell)
			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("Split(%q) err = %v, want %v", tc.input, err, tc.wantErr)
			}
			if !reflect.DeepEqual(words, tc.wantWords) {
				t.Fatalf("Split(%q) words = %q, want %q", tc.input, words, tc.wantWords)
			}
		})
	}
}

func TestAcceptMultiline(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{"complete", "echo hello", true},
		{"complete quoted", `echo "hello world"`, true},
		{"unterminated single quote", "echo 'oops", false},
		{"unterminated double quote", `echo "oops`, false},
		{"trailing backslash", `echo foo\`, false},
		{"empty", "", true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := AcceptMultiline([]rune(tc.input), EscapeShell); got != tc.want {
				t.Fatalf("AcceptMultiline(%q) = %v, want %v", tc.input, got, tc.want)
			}
		})
	}
}

func TestParseLiteral(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []string
	}{
		{"windows path", `ls C:\Windows\Temp`, []string{"ls", `C:\Windows\Temp`}},
		{"trailing backslash", `ls C:\Windows\Temp\`, []string{"ls", `C:\Windows\Temp\`}},
		{"escaped space kept literal", `echo a\ b`, []string{"echo", `a\`, "b"}},
		{"quotes still group", `echo "a b" c`, []string{"echo", "a b", "c"}},
		{"comment still stripped", `ls C:\Temp # note`, []string{"ls", `C:\Temp`}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := Parse(tc.input, EscapeLiteral)
			if err != nil {
				t.Fatalf("Parse(%q, literal): unexpected error: %v", tc.input, err)
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("Parse(%q, literal) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}

func TestAcceptMultilineLiteral(t *testing.T) {
	// A trailing backslash must never request another line in literal mode,
	// but unterminated quotes still do.
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{"trailing backslash accepted", `ls C:\Temp\`, true},
		{"windows path accepted", `ls C:\Windows\Temp`, true},
		{"unterminated single still waits", "echo 'oops", false},
		{"unterminated double still waits", `echo "oops`, false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := AcceptMultiline([]rune(tc.input), EscapeLiteral); got != tc.want {
				t.Fatalf("AcceptMultiline(%q, literal) = %v, want %v", tc.input, got, tc.want)
			}
		})
	}
}

func TestIsEmpty(t *testing.T) {
	empty := []rune{' ', '\t'}

	tests := []struct {
		name  string
		input string
		chars []rune
		want  bool
	}{
		{"empty string", "", empty, true},
		{"only spaces", "     ", empty, true},
		{"spaces and tabs", " \t \t ", empty, true},
		{"has content", "  x  ", empty, false},
		{"content no chars", "abc", nil, false},
		{"newline not in set", "\n", empty, false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := IsEmpty(tc.input, tc.chars...); got != tc.want {
				t.Fatalf("IsEmpty(%q) = %v, want %v", tc.input, got, tc.want)
			}
		})
	}
}

func TestTrimSpaces(t *testing.T) {
	got := TrimSpaces([]string{"  a  ", "b\t", "\tc"})
	want := []string{"a", "b", "c"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("TrimSpaces = %q, want %q", got, want)
	}
}

func TestUnescapeValue(t *testing.T) {
	tests := []struct {
		name       string
		prefixLine string
		val        string
		want       string
	}{
		{"double-quoted unescapes spaces", `"foo`, `bar\ baz`, "bar baz"},
		{"single-quoted unescapes spaces", `'foo`, `bar\ baz`, "bar baz"},
		{"unquoted left as-is", "foo", `bar\ baz`, `bar\ baz`},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := UnescapeValue("", tc.prefixLine, tc.val); got != tc.want {
				t.Fatalf("UnescapeValue(%q, %q) = %q, want %q", tc.prefixLine, tc.val, got, tc.want)
			}
		})
	}
}
