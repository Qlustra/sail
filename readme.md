Sail
====

## Population

`Explore[C]()` compiles the chart declaration once. `Populate[S](world)` then
returns a fresh `*C` populated against **S's type**, without requiring a live S
instance. Declared capabilities use zero-valued surrogates. The world contains
reusable schema definitions; populated markers
hold the metadata and subject field references for individual occurrences.
Populated markers expose `SubjectType()` and a copied `SubjectPath()` so consumers
can interpret receiving types and retain declaration order across spot kinds.

Population reads marked subject fields in declaration order, selecting spots
within the current lane. An unmatched field advances to the next lane only
after the current lane's minimum counts are satisfied. A matched spot exceeding
its maximum fails immediately; population does not retry another lane to make
the counts fit. End of declaration validates every remaining minimum.

Features receive text from their named tags. Qualifiers receive the remaining
mark values after discriminant selection. Untagged fields are ignored; marked
fields must be exported and match a remaining spot. Charts and regions apply
to structs (including through pointers); landmarks describe leaf declarations.

Single spots hold one marker, optional spots hold a marker pointer or nil, and
multi spots hold a slice. Repeated subject types are populated independently.
An active cycle of the same marker and subject types is rejected because it
would require an infinite expansion. Live subject binding remains separate.

## Barrels

A barrel declares a capability, selects a method, and stores its result during
population:

```go
type Descriptive interface {
	About() string
}

type CommandRegion struct {
	sail.Region
	Description sail.Barrel[string, Descriptive] `sail:"barrel" populate:"About"`
}

func (CommandRegion) Family() string { return "command" }
```

The capability must be an interface. Its selected method must be exported, take
no arguments, and return exactly one value assignable to the barrel's value
type. Exploration validates this contract before any subject is populated.

Omitting `barrel` means `barrel:"optional"`: an unavailable capability leaves
the barrel empty. `barrel:"required"` instead makes absence a population error.
Availability requires a concrete receiving type whose pointer implements the
interface; this covers value and pointer methods. An interface-typed subject
field cannot supply a concrete surrogate from its static declaration alone.

After the occurrence's fields and child lanes are populated, each barrel calls
its selected method on a fresh zero-valued surrogate. Hooks must work without
initialized application state. A hook panic fails population with the subject
path and barrel field in the error. Multiple barrels may select the same method;
they invoke independently and each stores its result.

`Description.Get()` reads the stored value. `Description.Present()` distinguishes
an absent capability from a successful call returning `""`, zero, or nil. Neither
accessor invokes hooks or validates values. No separate behavior registration or
population hook is required.

## Lane policies

Lanes form ordered bands in a subject's declaration. Spots are unordered
alternatives within a lane. The `exploration` tag declares how fields are
assigned to lanes:

| Policy | Meaning |
| --- | --- |
| `eager` | Stay in the current lane while it accepts the field. |
| `unambiguous` | Reject structural overlap between reachable lanes during exploration. |

Omitting `exploration` selects `unambiguous`.

Exploration rejects an unambiguous lane whose marks overlap a following lane
reachable through zero-occurrence lanes. A lane permits zero occurrences only
when every spot's minimum is zero. Lookahead includes the first required lane,
then stops. Marks are compared by tag, family, and discriminants; features and
qualifiers do not distinguish them. Ambiguous spots within one lane are rejected
under every policy.

Cardinality remains a separate population constraint. Finite maximum counts do
not resolve structural overlap, and satisfying counts does not excuse violating
the declared lane strategy.
