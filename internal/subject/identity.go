package subject

import (
	"fmt"
	"reflect"

	"github.com/qlustra/sail/internal/reflectutils"
)

// Identity is an opaque, comparable receiver identity within one propagation.
// Reference receivers use their actual storage. Value copies and inline
// zero-sized fields use occurrence provenance, without fabricated addresses.
// Identities should only be compared within the propagation that supplied them.
type Identity struct {
	typ     reflect.Type
	address uintptr
	origin  *valueOrigin
}

// Nonzero size keeps distinct provenance tokens distinct under Go's allocation
// rules. Inline fields reuse a canonical token for their containing identity.
type valueOrigin struct{ token byte }
type fieldOrigin struct {
	parent Identity
	index  int
}

type arrival struct {
	identity Identity
	first    bool
	cycle    *Encounter
}

func (p *propagator) identify(f *frame, value reflect.Value) (arrival, error) {
	signal, err := reflectutils.ReceiverIdentity(value)
	// The explicitly supplied root has unique invocation provenance even when
	// its pointer target is zero-sized. Entered pointers retain strict checks.
	if err != nil && f.parent == nil {
		terminal := value
		for terminal.IsValid() && (terminal.Kind() == reflect.Pointer || terminal.Kind() == reflect.Interface) && !terminal.IsNil() {
			terminal = terminal.Elem()
		}
		if terminal.IsValid() && terminal.Type().Size() == 0 {
			signal, err = reflectutils.ReceiverIdentity(terminal)
		}
	}
	if err != nil {
		return arrival{}, err
	}
	identity := Identity{typ: signal.Type, address: signal.Address}
	p.keep = append(p.keep, signal.Keep)
	if signal.Value {
		chain, present, err := f.navigator.CarrierChain(f.occurrence)
		if err != nil {
			return arrival{}, err
		}
		if !present {
			return arrival{}, fmt.Errorf("carrier disappeared while resolving identity")
		}
		identity, err = p.valueIdentity(f, chain)
		if err != nil {
			return arrival{}, err
		}
	}
	first, recorded := f.arrivals[identity]
	if !recorded {
		_, seen := p.seen[identity]
		first = !seen
		p.seen[identity] = struct{}{}
		f.arrivals[identity] = first
	}
	facts := arrival{identity: identity, first: first}
	// Ancestor scopes retain their pre-descent binding. Later field replacement
	// affects new arrivals, not the identity of a scope already in progress.
	for ancestor := f.parent; ancestor != nil; ancestor = ancestor.parent {
		if ancestor.current.identity == identity {
			at := ancestor.encounter()
			facts.cycle = &at
			break
		}
	}
	f.current = facts
	return facts, nil
}

func (p *propagator) valueIdentity(f *frame, chain []reflect.Value) (Identity, error) {
	var identity Identity
	for index, value := range chain {
		signal, err := reflectutils.ReceiverIdentity(value)
		if err != nil && f.parent == nil && index == 0 {
			for value.Kind() == reflect.Pointer || value.Kind() == reflect.Interface {
				value = value.Elem()
			}
			if value.Type().Size() == 0 {
				signal, err = reflectutils.ReceiverIdentity(value)
			}
		}
		if err != nil {
			return Identity{}, err
		}
		p.keep = append(p.keep, signal.Keep)
		if !signal.Value {
			identity = Identity{typ: signal.Type, address: signal.Address}
			continue
		}
		origin := f.origin
		if index > 0 {
			field := fieldOrigin{parent: identity, index: f.occurrence.Source.Path[index-1]}
			origin = p.origins[field]
			if origin == nil {
				origin = &valueOrigin{}
				p.origins[field] = origin
			}
		}
		identity = Identity{typ: signal.Type, origin: origin}
	}
	return identity, nil
}
