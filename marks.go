package sail

import (
	"fmt"
	"slices"
	"strings"
)

// mark identifies one descriptor within a tag's mark family. It owns its
// normalized discriminator slice; declarations may safely reuse their input.
type mark struct {
	tag           string
	family        string
	discriminants []string
}

type markIdentity struct {
	tag, family, discriminants string
}

func (m mark) identity() markIdentity {
	return markIdentity{tag: m.tag, family: m.family, discriminants: strings.Join(m.discriminants, ",")}
}

func (m mark) String() string {
	value := m.family
	if len(m.discriminants) > 0 {
		value += "," + strings.Join(m.discriminants, ",")
	}
	return fmt.Sprintf("%s:%q", m.tag, value)
}

type markFamily struct {
	tag    string
	family string
}

// Discriminator vocabulary is global to a family, even where only some of the
// family's descriptors are eligible. This prevents an unavailable variant from
// being interpreted as an ordinary mark with extra qualifiers.
type markFamilies map[markFamily]map[string]struct{}

func newMark(tag, family string, discriminants []string) (mark, error) {
	if err := validateTagKey(tag); err != nil {
		return mark{}, err
	}
	if err := validateTagComponent(family, true); err != nil {
		return mark{}, fmt.Errorf("invalid mark family: %w", err)
	}

	m := mark{
		tag:           tag,
		family:        family,
		discriminants: slices.Clone(discriminants),
	}
	slices.Sort(m.discriminants)
	for i, discriminator := range m.discriminants {
		if err := validateTagComponent(discriminator, false); err != nil {
			return mark{}, fmt.Errorf("invalid mark discriminator: %w", err)
		}
		if i > 0 && m.discriminants[i-1] == discriminator {
			return mark{}, fmt.Errorf("duplicate mark discriminator %q", discriminator)
		}
	}
	return m, nil
}

// compileMarkFamilies runs after graph discovery, when every reachable
// descriptor has contributed its discriminator vocabulary.
func compileMarkFamilies(descriptors []*markerDescriptor) (markFamilies, error) {
	families := make(markFamilies)
	for _, descriptor := range descriptors {
		if descriptor == nil {
			return nil, fmt.Errorf("cannot compile a nil descriptor")
		}
		family := markFamily{tag: descriptor.mark.tag, family: descriptor.mark.family}
		vocabulary := families[family]
		if vocabulary == nil {
			vocabulary = make(map[string]struct{})
			families[family] = vocabulary
		}
		for _, discriminator := range descriptor.mark.discriminants {
			vocabulary[discriminator] = struct{}{}
		}
	}

	for _, descriptor := range descriptors {
		family := markFamily{tag: descriptor.mark.tag, family: descriptor.mark.family}
		for _, qualifier := range descriptor.qualifiers {
			if _, conflict := families[family][qualifier.symbol]; conflict {
				return nil, fmt.Errorf("descriptor %v: qualifier %q conflicts with a discriminator in mark family %q/%q", descriptor.typ, qualifier.symbol, family.tag, family.family)
			}
		}
	}
	return families, nil
}
