package definitions

import (
	"github.com/qlustra/intel/define"
	"github.com/qlustra/intel/logic"
	"github.com/qlustra/intel/rules/property"
)

type CarrierEncounterHandler[
	HOST define.ScopeDeclaration,
] func(property.Encounter, *HOST) logic.Check

type CarrierApi[HOST define.ScopeDeclaration] interface {
	Indicator() *property.Rule
	Host(encounter property.Encounter) (*HOST, logic.Check)
	Encounter(interceptor CarrierEncounterHandler[HOST])
	Handle(encounter property.Encounter) logic.Check
}

type CarrierDefinition[
	HOST define.ScopeDeclaration,
	INDICATOR define.LeafIndicator,
] struct {
	HostMark    string
	interceptor CarrierEncounterHandler[HOST]
}

func Carrier[
	HOST define.ScopeDeclaration,
	INDICATOR define.LeafIndicator,
](hostMark string) *CarrierDefinition[HOST, INDICATOR] {
	return &CarrierDefinition[HOST, INDICATOR]{
		HostMark: hostMark,
	}
}

func (c *CarrierDefinition[HOST, INDICATOR]) Indicator() *property.Rule {
	return property.Satisfies[INDICATOR]()
}

func (c *CarrierDefinition[HOST, INDICATOR]) Host(encounter property.Encounter) (*HOST, logic.Check) {
	return encounter.Access[HOST](c.HostMark)
}

func (c *CarrierDefinition[HOST, INDICATOR]) Encounter(interceptor CarrierEncounterHandler[HOST]) {
	c.interceptor = interceptor
}

func (c *CarrierDefinition[HOST, INDICATOR]) Handle(encounter property.Encounter) logic.Check {
	host, check := c.Host(encounter)
	if !check.OK() {
		return check
	}
	if c.interceptor == nil {
		return logic.Ok()
	}
	return c.interceptor(encounter, host)
}
