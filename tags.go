package sail

import (
	"fmt"
	"reflect"
	"strings"
	"unicode"
)

// TagValidator validates one parsed value without inspecting a declaration or
// consulting its symbol ledger.
type TagValidator interface {
	Validate() error
}

type TagFormat interface {
	TagValidator
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

type EmptyTag struct {
	Tag
}

func (t EmptyTag) Validate() error {
	if t.Raw != "" {
		return fmt.Errorf("expected an empty tag value, got %q", t.Raw)
	}
	return nil
}

type SingleValueTag struct {
	Tag
	// Validator is optional; unrestricted values such as help text are valid
	// unless a caller supplies a more specific rule.
	Validator func(string) error
}

func (t SingleValueTag) Validate() error {
	if t.Validator != nil {
		return t.Validator(t.Raw)
	}
	return nil
}

type CommaValueTag struct {
	Tag
	Parts     []string
	Validator func([]string) error
}

func (t *CommaValueTag) Parse(raw string) {
	t.Raw = raw
	t.Parts = strings.Split(raw, ",")
}

func (t CommaValueTag) Validate() error {
	if t.Validator != nil {
		return t.Validator(t.Parts)
	}
	return nil
}

// readTag keeps named lookup separate from value parsing and validation. A
// present empty value is parsed; a missing tag returns false without parsing.
// Callers use a fresh format for each read and retain only its parsed data.
func readTag(tags reflect.StructTag, name string, format TagFormat) (bool, error) {
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

func readSingleTag(tags reflect.StructTag, name string, validator func(string) error) (string, error) {
	format := SingleValueTag{Tag: Tag{Name: name}, Validator: validator}
	present, err := readTag(tags, name, &format)
	if err != nil {
		return "", err
	}
	if !present {
		return "", fmt.Errorf("missing tag %q", name)
	}
	return format.Raw, nil
}

func readSailTag(tags reflect.StructTag) (CommaValueTag, bool, error) {
	format := CommaValueTag{Tag: Tag{Name: "sail"}, Validator: validateSailTag}
	present, err := readTag(tags, "sail", &format)
	return format, present, err
}

func validateSailTag(parts []string) error {
	if len(parts) == 0 {
		return fmt.Errorf("missing Sail role")
	}
	if err := validateTagComponent(parts[0], false); err != nil {
		return fmt.Errorf("invalid Sail role: %w", err)
	}
	if err := validateTagValues(parts[1:]); err != nil {
		return err
	}
	// Comma tails are parsed, but Sail does not define declaration qualifiers yet.
	if len(parts) > 1 {
		return fmt.Errorf("unsupported Sail qualifiers %q", parts[1:])
	}
	return nil
}

func validateSymbol(value string) error {
	if value == "" || strings.ContainsAny(value, ",:\"") || strings.ContainsFunc(value, invalidTagRune) {
		return fmt.Errorf("invalid symbol %q", value)
	}
	return nil
}

func validateTagKey(value string) error {
	if value == "" || strings.ContainsAny(value, "\":") || strings.ContainsFunc(value, invalidTagRune) {
		return fmt.Errorf("invalid tag key %q", value)
	}
	return nil
}

func invalidTagRune(r rune) bool {
	return unicode.IsSpace(r) || unicode.IsControl(r)
}

func validateTagComponent(value string, allowEmpty bool) error {
	if (!allowEmpty && value == "") || strings.ContainsRune(value, ',') || strings.ContainsFunc(value, invalidTagRune) {
		return fmt.Errorf("invalid tag component %q", value)
	}
	return nil
}

// validateTagValues checks trailing comma components independently of the
// leading role or family. Their meaning is resolved against the chart later.
func validateTagValues(values []string) error {
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		if err := validateTagComponent(value, false); err != nil {
			return err
		}
		if _, duplicate := seen[value]; duplicate {
			return fmt.Errorf("duplicate tag component %q", value)
		}
		seen[value] = struct{}{}
	}
	return nil
}

// parsedMarkTag is local to one surveyed application field. Compiled descriptors
// never retain mutable tag parser instances or application values.
type parsedMarkTag struct {
	family string
	values []string
}

func parseMarkTag(raw string) (parsedMarkTag, error) {
	format := CommaValueTag{Validator: validateMarkTag}
	format.Parse(raw)
	if err := format.Validate(); err != nil {
		return parsedMarkTag{}, err
	}
	return parsedMarkTag{family: format.Parts[0], values: format.Parts[1:]}, nil
}

func validateMarkTag(parts []string) error {
	if len(parts) == 0 {
		return fmt.Errorf("missing mark family")
	}
	if err := validateTagComponent(parts[0], true); err != nil {
		return fmt.Errorf("invalid mark family: %w", err)
	}
	return validateTagValues(parts[1:])
}
