package tag

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
)

var (
	// ErrInvalidTag indicates an invalid tag or invalid use of an existing tag.
	ErrInvalidTag = errors.New("invalid tag")

	// ErrTag indicates an error while parsing flag tags.
	ErrTag = errors.New("tag error")
)

// simple wrapper for errors.
func newError(err error, msg string) error {
	return fmt.Errorf("%w: %s", err, msg)
}

// MultiTag is a structure to efficiently query a
// struct field' tags, regardless of their complexity.
type MultiTag struct {
	value string
	cache map[string][]string
}

// NewMultiTag returns a new multi tag from a field tag string.
// The tags have not been parsed, you must call tag.Parse().
func NewMultiTag(v string) MultiTag {
	return MultiTag{
		value: v,
	}
}

// GetFieldTag directly reads, creates and parses a struct field and
// returns either a working multiTag, or an error if parsing failed.
func GetFieldTag(field reflect.StructField) (MultiTag, bool, error) {
	var err error

	// PkgName is set only for non-exported fields, which we ignore
	if field.PkgPath != "" && !field.Anonymous {
		return MultiTag{}, true, nil
	}

	// If the field tag is empty, there is no tag
	if field.Tag == "" {
		return MultiTag{}, true, nil
	}

	// Else find the tag and try parsing
	mtag := NewMultiTag(string(field.Tag))

	// 1 - with our own code, for multiple tags.
	if err = mtag.Parse(); err != nil {
		return mtag, false, err
	}

	// Skip fields with the no-flag tag
	if noFlag, _ := mtag.Get("no-flag"); noFlag != "" {
		return mtag, true, nil
	}

	return mtag, false, nil
}

// Parse scans the struct tag string for all keys and their values.
func (x *MultiTag) Parse() error {
	vals, err := x.scan()
	x.cache = vals

	return err
}

// Get returns the value of a key, and if this key is set.
func (x *MultiTag) Get(key string) (string, bool) {
	c := x.cached()

	if v, ok := c[key]; ok {
		return v[len(v)-1], true
	}

	return "", false
}

// Get returns the values of a key, and if this key is set.
func (x *MultiTag) GetMany(key string) []string {
	c := x.cached()

	return c[key]
}

// Set changes the value of a key in the cache.
func (x *MultiTag) Set(key string, value string) {
	c := x.cached()
	c[key] = []string{value}
}

// SetMany stores some values in the cache, for the given key.
func (x *MultiTag) SetMany(key string, value []string) {
	c := x.cached()
	c[key] = value
}

func (x *MultiTag) scan() (map[string][]string, error) {
	val := x.value

	ret := make(map[string][]string)

	// This is mostly copied from reflect.StructTag.Get
	for val != "" {
		pos := 0

		// Skip whitespace
		for pos < len(val) && val[pos] == ' ' {
			pos++
		}

		val = val[pos:]

		if val == "" {
			break
		}

		// Scan to colon to find key
		name, pos, kerr := x.scanForKey(val)
		if kerr != nil {
			return nil, kerr
		}

		val = val[pos+1:]

		// Scan quoted string to find value
		value, pos, verr := x.scanForValue(val, name)
		if verr != nil {
			return nil, verr
		}

		val = val[pos+1:]

		ret[name] = append(ret[name], value)
	}

	return ret, nil
}

func (x *MultiTag) scanForKey(val string) (string, int, error) {
	pos := 0

	for pos < len(val) && val[pos] != ' ' && val[pos] != ':' && val[pos] != '"' {
		pos++
	}

	if kerr := x.keyError(pos, val); kerr != nil {
		return "", pos, kerr
	}

	return val[:pos], pos, nil
}

func (x *MultiTag) scanForValue(val string, name string) (string, int, error) {
	pos := 1

	for pos < len(val) && val[pos] != '"' {
		if val[pos] == '\n' {
			msg := fmt.Sprintf("unexpected newline in tag value `%v' (in `%v`)", name, x.value)

			return "", pos, newError(ErrTag, msg)
		}

		if val[pos] == '\\' {
			pos++
		}
		pos++
	}

	if pos >= len(val) {
		msg := fmt.Sprintf("expected end of tag value `\"' at end of tag (in `%v`)", x.value)

		return "", pos, newError(ErrTag, msg)
	}

	value, err := strconv.Unquote(val[:pos+1])
	if err != nil {
		msg := fmt.Sprintf("Malformed value of tag `%v:%v` => %v (in `%v`)", name, val[:pos+1], err, x.value)

		return "", pos, newError(ErrTag, msg)
	}

	return value, pos, nil
}

func (x *MultiTag) keyError(index int, val string) error {
	if index >= len(val) {
		msg := fmt.Sprintf("expected `:' after key name, but got end of tag (in `%v`)", x.value)

		return newError(ErrTag, msg)
	}

	if val[index] != ':' {
		msg := fmt.Sprintf("expected `:' after key name, but got `%v' (in `%v`)", val[index], x.value)

		return newError(ErrTag, msg)
	}

	if index+1 >= len(val) {
		msg := fmt.Sprintf("expected `\"' to start tag value at end of tag (in `%v`)", x.value)

		return newError(ErrTag, msg)
	}

	if val[index+1] != '"' {
		msg := fmt.Sprintf("expected `\"' to start tag value, but got `%v' (in `%v`)", val[index+1], x.value)

		return newError(ErrTag, msg)
	}

	return nil
}

func (x *MultiTag) cached() map[string][]string {
	if x.cache == nil {
		cache, _ := x.scan()

		if cache == nil {
			cache = make(map[string][]string)
		}

		x.cache = cache
	}

	return x.cache
}
