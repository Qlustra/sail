package engine

import (
	"fmt"
	"reflect"

	"github.com/qlustra/sail/internal/reflectutils"
)

// ValidateVessel checks the populated Chart against its Charter and the live
// Subject root. Descendants remain subject to binding when reached.
func ValidateVessel(charter *Charter, chart reflect.Value, subject any) error {
	root, err := navigationRoot(charter, chart)
	if err != nil {
		return err
	}
	value, err := bindReference(root.Source, subject)
	if err == nil {
		_, err = reflectutils.Dereference(value)
	}
	if err != nil {
		return fmt.Errorf("%s: %w", root.Source, err)
	}
	return nil
}
