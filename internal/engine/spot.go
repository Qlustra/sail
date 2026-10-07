package engine

import (
	"fmt"
	"reflect"

	decl "github.com/qlustra/sail/internal/declaration"
	"github.com/qlustra/sail/internal/reflectutils"
)

// validateSpotEmbeddings checks composition before invoking the promoted spot
// method on a zero-valued surrogate. Only embeddings that provide the spot
// declaration participate; populated Value fields are outside this policy.
func validateSpotEmbeddings(typ reflect.Type) error {
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		if !field.Anonymous || !reflectutils.FieldImplements[decl.SpotDeclaration](field) {
			continue
		}
		if field.Type.Kind() == reflect.Pointer {
			return fmt.Errorf("%v.%s: spot declarations must be embedded by value, not pointer", typ, field.Name)
		}
		if err := validateSpotEmbeddings(field.Type); err != nil {
			return fmt.Errorf("%v.%s: %w", typ, field.Name, err)
		}
	}
	return nil
}
