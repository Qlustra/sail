package sail

import "github.com/qlustra/sail/internal/chart/exploration"

// Compile validates the authored Legend struct L and returns its reusable
// Navigator. It does not validate any specific Chart. Nonfatal compilation
// advisories are available through Navigator.Warnings.
func Compile[L any]() (*Navigator[L], error) {
	explorer, err := exploration.New()
	if err != nil {
		return nil, err
	}
	legend, err := explorer.Explore[L]()
	if err != nil {
		return nil, err
	}
	return &Navigator[L]{grammar: legend}, nil
}
