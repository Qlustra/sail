package engine

import (
	"fmt"
	"reflect"
	"strings"
	"unicode"

	"github.com/qlustra/sail/internal/tags"
)

func readSailTag(structTags reflect.StructTag) (tags.CommaValue, bool, error) {
	format := tags.CommaValue{Tag: tags.Tag{Name: "sail"}, Validator: validateSailTag}
	present, err := tags.Read(structTags, "sail", &format)
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
// leading role or family. Their meaning is resolved against the charter later.
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

// parsedMarkTag is local to one surveyed subject field. Compiled descriptors
// never retain mutable tag parser instances or subject values.
type parsedMarkTag struct {
	family string
	values []string
}

func parseMarkTag(raw string) (parsedMarkTag, error) {
	format := tags.CommaValue{Validator: validateMarkTag}
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
