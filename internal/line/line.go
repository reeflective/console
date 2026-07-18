package line

import (
	"bytes"
	"errors"
	"strings"
	"unicode/utf8"

	"github.com/kballard/go-shellquote"
	"mvdan.cc/sh/v3/syntax"
)

var (
	SplitChars        = " \n\t"
	SingleChar        = '\''
	DoubleChar        = '"'
	EscapeChar        = '\\'
	DoubleEscapeChars = "$`\"\n\\"
)

var (
	ErrUnterminatedSingleQuote = errors.New("unterminated single-quoted string")
	ErrUnterminatedDoubleQuote = errors.New("unterminated double-quoted string")
	ErrUnterminatedEscape      = errors.New("unterminated backslash-escape")
)

// EscapeMode controls how the line parser treats backslashes when splitting an
// input line into words.
type EscapeMode int

const (
	// EscapeShell is the default POSIX-shell behaviour: a backslash escapes the
	// following character, so `C:\Windows` becomes `C:Windows`, and a trailing
	// backslash marks the line as an incomplete continuation.
	EscapeShell EscapeMode = iota

	// EscapeLiteral preserves backslashes as ordinary characters. Quotes still
	// group words and are removed, but `C:\Windows\Temp` is passed through
	// verbatim and a trailing backslash does not request another line.
	EscapeLiteral
)

// Parse is in charge of removing all comments from the input line
// before execution, and if successfully parsed, split into words.
//
// The mode governs how backslashes are treated when the (comment-stripped)
// line is split into words: EscapeShell applies POSIX escape rules, while
// EscapeLiteral preserves backslashes verbatim.
func Parse(input string, mode EscapeMode) (args []string, err error) {
	lineReader := strings.NewReader(input)
	parser := syntax.NewParser(syntax.KeepComments(false))

	// Parse the shell string a syntax, removing all comments.
	stmts, err := parser.Parse(lineReader, "")
	if err != nil {
		return nil, err
	}

	var parsedLine bytes.Buffer

	err = syntax.NewPrinter().Print(&parsedLine, stmts)
	if err != nil {
		return nil, err
	}

	// In literal mode, split with our own splitter so that backslashes (e.g. in
	// Windows paths) are preserved instead of being consumed as shell escapes.
	if mode == EscapeLiteral {
		args, _, err = Split(parsedLine.String(), false, EscapeLiteral)

		return args, err
	}

	// Split the line into shell words.
	return shellquote.Split(parsedLine.String())
}

// acceptMultiline determines if the line just accepted is complete (in which case
// we should execute it), or incomplete (in which case we must read in multiline).
//
// The mode controls escape handling: in EscapeLiteral, a trailing backslash is an
// ordinary character and never requests another line (only unterminated quotes do).
func AcceptMultiline(line []rune, mode EscapeMode) (accept bool) {
	// Errors are either: unterminated quotes, or unterminated escapes.
	_, _, err := Split(string(line), false, mode)
	if err == nil {
		return true
	}

	// Currently, unterminated quotes are obvious to treat: keep reading.
	switch err {
	case ErrUnterminatedDoubleQuote, ErrUnterminatedSingleQuote:
		return false
	case ErrUnterminatedEscape:
		if len(line) > 0 && line[len(line)-1] == '\\' {
			return false
		}

		return true
	}

	return true
}

// IsEmpty checks if a given input line is empty.
// It accepts a list of characters that we consider to be irrelevant,
// that is, if the given line only contains these characters, it will
// be considered empty.
func IsEmpty(line string, emptyChars ...rune) bool {
	empty := true

	for _, r := range line {
		if !strings.ContainsRune(string(emptyChars), r) {
			empty = false
			break
		}
	}

	return empty
}

// UnescapeValue is used When the completer has returned us some completions, 
// we sometimes need to post-process them a little before passing them to our shell.
func UnescapeValue(prefixComp, prefixLine, val string) string {
	quoted := strings.HasPrefix(prefixLine, "\"") ||
		strings.HasPrefix(prefixLine, "'")

	if quoted {
		val = strings.ReplaceAll(val, "\\ ", " ")
	}

	return val
}

// TrimSpaces removes all leading/trailing spaces from words
func TrimSpaces(remain []string) (trimmed []string) {
	for _, word := range remain {
		trimmed = append(trimmed, strings.TrimSpace(word))
	}

	return
}

// Split has been copied from go-shellquote and slightly modified so as to also
// return the remainder when the parsing failed because of an unterminated quote.
//
// In EscapeLiteral mode, backslashes are treated as ordinary characters: they
// are neither consumed as escapes nor able to mark a line continuation.
func Split(input string, hl bool, mode EscapeMode) (words []string, remainder string, err error) {
	var buf bytes.Buffer
	words = make([]string, 0)

	for len(input) > 0 {
		// skip any splitChars at the start
		c, l := utf8.DecodeRuneInString(input)
		if strings.ContainsRune(SplitChars, c) {
			// Keep these characters in the result when higlighting the line.
			if hl {
				if len(words) == 0 {
					words = append(words, string(c))
				} else {
					words[len(words)-1] += string(c)
				}
			}

			input = input[l:]

			continue
		} else if c == EscapeChar && mode == EscapeShell {
			// Look ahead for escaped newline so we can skip over it
			next := input[l:]
			if len(next) == 0 {
				if hl {
					remainder = string(EscapeChar)
				}

				err = ErrUnterminatedEscape

				return words, remainder, err
			}

			c2, l2 := utf8.DecodeRuneInString(next)
			if c2 == '\n' {
				if hl {
					if len(words) == 0 {
						words = append(words, string(c)+string(c2))
					} else {
						words[len(words)-1] += string(c) + string(c2)
					}
				}

				input = next[l2:]

				continue
			}
		}

		var word string

		word, input, err = splitWord(input, &buf, hl, mode)
		if err != nil {
			remainder = input
			return words, remainder, err
		}

		words = append(words, word)
	}

	return words, remainder, err
}

// splitWord has been modified to return the remainder of the input (the part that has not been
// added to the buffer) even when an error is returned.
func splitWord(input string, buf *bytes.Buffer, hl bool, mode EscapeMode) (word string, remainder string, err error) {
	buf.Reset()

raw:
	{
		cur := input
		for len(cur) > 0 {
			c, l := utf8.DecodeRuneInString(cur)
			cur = cur[l:]
			if c == SingleChar {
				buf.WriteString(input[0 : len(input)-len(cur)-l])
				input = cur
				goto single
			} else if c == DoubleChar {
				buf.WriteString(input[0 : len(input)-len(cur)-l])
				input = cur
				goto double
			} else if c == EscapeChar && mode == EscapeShell {
				buf.WriteString(input[0 : len(input)-len(cur)-l])
				if hl {
					buf.WriteRune(c)
				}
				input = cur
				goto escape
			} else if strings.ContainsRune(SplitChars, c) {
				buf.WriteString(input[0 : len(input)-len(cur)-l])
				if hl {
					buf.WriteRune(c)
				}

				return buf.String(), cur, nil
			}
		}
		if len(input) > 0 {
			buf.WriteString(input)
			input = ""
		}
		goto done
	}

escape:
	{
		if len(input) == 0 {
			if hl {
				input = buf.String() + input
			}
			return "", input, ErrUnterminatedEscape
		}
		c, l := utf8.DecodeRuneInString(input)
		if c == '\n' {
			// a backslash-escaped newline is elided from the output entirely
		} else {
			buf.WriteString(input[:l])
		}
		input = input[l:]
	}

	goto raw

single:
	{
		i := strings.IndexRune(input, SingleChar)
		if i == -1 {
			if hl {
				input = buf.String() + YellowFG + string(SingleChar) + input
			}
			return "", input, ErrUnterminatedSingleQuote
		}
		// Catch up opening quote
		if hl {
			buf.WriteString(YellowFG)
			buf.WriteRune(SingleChar)
		}

		buf.WriteString(input[0:i])
		input = input[i+1:]

		if hl {
			buf.WriteRune(SingleChar)
			buf.WriteString(ResetFG)
		}
		goto raw
	}

double:
	{
		cur := input
		for len(cur) > 0 {
			c, l := utf8.DecodeRuneInString(cur)
			cur = cur[l:]
			if c == DoubleChar {
				// Catch up opening quote
				if hl {
					buf.WriteString(YellowFG)
					buf.WriteRune(c)
				}

				buf.WriteString(input[0 : len(input)-len(cur)-l])

				if hl {
					buf.WriteRune(c)
					buf.WriteString(ResetFG)
				}
				input = cur
				goto raw
			} else if c == EscapeChar && !hl && mode == EscapeShell {
				// bash only supports certain escapes in double-quoted strings
				c2, l2 := utf8.DecodeRuneInString(cur)
				cur = cur[l2:]
				if strings.ContainsRune(DoubleEscapeChars, c2) {
					buf.WriteString(input[0 : len(input)-len(cur)-l-l2])
					if c2 == '\n' {
						// newline is special, skip the backslash entirely
					} else {
						buf.WriteRune(c2)
					}
					input = cur
				}
			}
		}

		if hl {
			input = buf.String() + YellowFG + string(DoubleChar) + input
		}

		return "", input, ErrUnterminatedDoubleQuote
	}

done:
	return buf.String(), input, nil
}

