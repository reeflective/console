package completion

import (
	"reflect"
	"testing"

	"github.com/reeflective/console/internal/line"
)

func TestSplitCompWords(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		wantWords     []string
		wantRemainder string
		wantErr       error
	}{
		{"empty", "", []string{}, "", nil},
		{"two words", "echo hello", []string{"echo", "hello"}, "", nil},
		{"single quoted", "echo 'hello world'", []string{"echo", "hello world"}, "", nil},
		{"double quoted", `echo "hello world"`, []string{"echo", "hello world"}, "", nil},
		{"unterminated single", "echo 'foo", []string{"echo"}, "foo", line.ErrUnterminatedSingleQuote},
		{"unterminated double", `echo "foo`, []string{"echo"}, "foo", line.ErrUnterminatedDoubleQuote},
		{"trailing backslash", `echo foo\`, []string{"echo"}, `foo\`, line.ErrUnterminatedEscape},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			words, remainder, err := splitCompWords(tc.input)
			if err != tc.wantErr {
				t.Fatalf("splitCompWords(%q) err = %v, want %v", tc.input, err, tc.wantErr)
			}
			if !reflect.DeepEqual(words, tc.wantWords) {
				t.Fatalf("splitCompWords(%q) words = %q, want %q", tc.input, words, tc.wantWords)
			}
			if remainder != tc.wantRemainder {
				t.Fatalf("splitCompWords(%q) remainder = %q, want %q", tc.input, remainder, tc.wantRemainder)
			}
		})
	}
}

func TestAdjustQuotedPrefix(t *testing.T) {
	tests := []struct {
		name      string
		remain    string
		err       error
		wantArg   string
		wantComp  string
		wantInput string
	}{
		{"no error", "foo", nil, "foo", "", ""},
		{"double quote", "foo", line.ErrUnterminatedDoubleQuote, "foo", `"`, `"foo`},
		{"single quote", "foo", line.ErrUnterminatedSingleQuote, "foo", "'", "'foo"},
		{"escape strips backslashes", `fo\o`, line.ErrUnterminatedEscape, "foo", "", ""},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			arg, comp, input := adjustQuotedPrefix(tc.remain, tc.err)
			if arg != tc.wantArg || comp != tc.wantComp || input != tc.wantInput {
				t.Fatalf("adjustQuotedPrefix(%q) = (%q, %q, %q), want (%q, %q, %q)",
					tc.remain, arg, comp, input, tc.wantArg, tc.wantComp, tc.wantInput)
			}
		})
	}
}

func TestSanitizeArgs(t *testing.T) {
	got := sanitizeArgs([]string{"a\nb", "c\td", `e\ f`})
	want := []string{"a b", "c d", "e f"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("sanitizeArgs = %q, want %q", got, want)
	}
}

func TestSplitArgs(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		wantArgs       []string
		wantPrefixComp string
		wantPrefixLine string
	}{
		{"empty line completes root", "", []string{""}, "", ""},
		{"partial word", "cmd", []string{"cmd"}, "", ""},
		{"trailing space starts new word", "cmd ", []string{"cmd", ""}, "", ""},
		{"two words", "cmd arg", []string{"cmd", "arg"}, "", ""},
		{"unterminated double quote", `cmd "foo`, []string{"cmd", "foo"}, `"`, `"foo`},
		{"unterminated single quote", "cmd 'foo", []string{"cmd", "foo"}, "'", "'foo"},
		{"color codes stripped", "\x1b[32mcmd\x1b[0m", []string{"cmd"}, "", ""},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			runes := []rune(tc.input)
			args, prefixComp, prefixLine := SplitArgs(runes, len(runes))
			if !reflect.DeepEqual(args, tc.wantArgs) {
				t.Fatalf("SplitArgs(%q) args = %q, want %q", tc.input, args, tc.wantArgs)
			}
			if prefixComp != tc.wantPrefixComp {
				t.Fatalf("SplitArgs(%q) prefixComp = %q, want %q", tc.input, prefixComp, tc.wantPrefixComp)
			}
			if prefixLine != tc.wantPrefixLine {
				t.Fatalf("SplitArgs(%q) prefixLine = %q, want %q", tc.input, prefixLine, tc.wantPrefixLine)
			}
		})
	}
}
