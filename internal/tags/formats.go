package tags

import (
	"fmt"
	"reflect"
	"strings"
)

// Validator validates one parsed value without inspecting a declaration or
// consulting its symbol ledger.
type Validator interface {
	Validate() error
}

type Format interface {
	Validator
	Recognize(name string) bool
	Parse(raw string)
}

type Tag struct {
	Name string
	Raw  string
}

func (t Tag) Recognize(name string) bool {
	return t.Name == name
}

func (t *Tag) Parse(raw string) {
	t.Raw = raw
}

// Validate accepts opaque tag text. Specific formats supply their own rules.
func (t Tag) Validate() error {
	return nil
}

type Empty struct {
	Tag
}

func (t Empty) Validate() error {
	if t.Raw != "" {
		return fmt.Errorf("expected an empty tag value, got %q", t.Raw)
	}
	return nil
}

type SingleValue struct {
	Tag
	// Validator is optional; unrestricted values such as help text are valid
	// unless a caller supplies a more specific rule.
	Validator func(string) error
}

func (t SingleValue) Validate() error {
	if t.Validator != nil {
		return t.Validator(t.Raw)
	}
	return nil
}

type CommaValue struct {
	Tag
	Parts     []string
	Validator func([]string) error
}

func (t *CommaValue) Parse(raw string) {
	t.Raw = raw
	t.Parts = strings.Split(raw, ",")
}

func (t CommaValue) Validate() error {
	if t.Validator != nil {
		return t.Validator(t.Parts)
	}
	return nil
}

// Read keeps named lookup separate from value parsing and validation. A
// present empty value is parsed; a missing tag returns false without parsing.
// Callers use a fresh format for each read and retain only its parsed data.
func Read(tags reflect.StructTag, name string, format Format) (bool, error) {
	raw, present := tags.Lookup(name)
	if !present {
		return false, nil
	}
	if !format.Recognize(name) {
		return true, fmt.Errorf("tag format does not recognize %q", name)
	}
	format.Parse(raw)
	if err := format.Validate(); err != nil {
		return true, fmt.Errorf("tag %q: %w", name, err)
	}
	return true, nil
}

func ReadSingle(tags reflect.StructTag, name string, validator func(string) error) (string, error) {
	format := SingleValue{Tag: Tag{Name: name}, Validator: validator}
	present, err := Read(tags, name, &format)
	if err != nil {
		return "", err
	}
	if !present {
		return "", fmt.Errorf("missing tag %q", name)
	}
	return format.Raw, nil
}
