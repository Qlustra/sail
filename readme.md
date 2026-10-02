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

## Acceptance

A marker can constrain its receiving type with an optional declaration hook:

```go
func (EnumOptionLandmark) Accept() sail.Acceptance {
	return sail.Exact[bool]()
}
```

Exploration invokes `Accept` once per location on its zero-valued surrogate and
records the returned rule. Returning an empty `sail.Acceptance` or panicking in
the hook fails exploration. Without the hook, existing role requirements apply.

Population checks the original declared subject type before filling that
occurrence's metadata, descending, or invoking its capabilities. Exact matching
accepts aliases of T, but not distinct defined types or pointers to T. Interface
types also match by identity: `Exact[any]()` requires a field declared `any`.
Charts and regions must still resolve to structs.

Marks and lanes select the occurrence first. An acceptance mismatch fails with
the subject path, marker, expected type, and actual type; it does not retry
another spot or advance the lane. Only selected occurrences are checked;
untagged fields remain outside population.

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

## Qualities

A qualifier composite can declare capabilities required when its symbol is
present on a subject occurrence:

```go
type Defaultable[T any] interface {
	Default() T
}

type DefaultQualifier struct {
	sail.Qualifier
	Value sail.Quality[any, Defaultable[sail.Self]] `sail:"quality" populate:"Default"`
}

type Argument struct {
	sail.Landmark
	Default DefaultQualifier `sail:"qualifier" symbol:"default"`
}

func (Argument) Family() string { return "arg" }

type Remote string

func (Remote) Default() Remote { return "origin" }
```

For a subject field `Remote Remote` marked with `arg,default`, population
requires and invokes `Default() Remote`. Without the qualifier, the quality
stays inactive: no capability is required or invoked. This association comes
from nesting, not from the symbol's meaning. Reusable qualifier wrappers retain
the qualities declared along their embedded `Qualifier` path.

`Self` matches the receiving type after pointer normalization, preserving
defined-type identity. It is a signature placeholder, not a Go interface that
the subject must literally implement. Sail checks the entire capability's
method set. This slice supports Self directly in method parameters and results;
compound uses such as `[]Self` and `*Self` fail exploration. Interfaces using
Self must have exported methods. Ordinary interfaces use Go's implementation
rules, including private methods.

The selected `populate` method must be exported, take no arguments, and return
one value. Fixed result types are checked for assignability to the storage type
during exploration; Self results are checked during population. Self cannot be
used as the storage type. With `any`, `arg.Default.Value.Get()` returns an
interface containing the actual `Remote` value, not a converted string.

Each active quality invokes its method on a fresh zero-valued surrogate, after
the occurrence's fields and child lanes are populated. Missing or mismatched
capabilities, incompatible storage, and hook panics fail population with the
subject and qualifier context. `Present()` distinguishes an inactive quality
from a successful zero or nil result. Reading a quality never invokes or
validates its value. Live-value parsing and validation remain separate.

## Declaration utilities

`ResolvePointerChain(typ)` follows pointer types without erasing the identity of
the resulting defined type. Nil input or a pointer cycle returns nil. An
interface remains an interface; a declaration alone cannot resolve its concrete
implementation.

`ReadCapability[V, I](typ, method)` reads a declaration capability without a
chart or qualifier. It uses the same contract matching and invocation machinery
as qualities:

```go
value, present, err := sail.ReadCapability[any, Defaultable[sail.Self]](
	reflect.TypeFor[Remote](), "Default",
)
```

I must be an interface. Its selected method must be exported, take no arguments,
and return one value assignable to V. Self follows the same direct-signature
rules as qualities. In this example, `value` contains a `Remote`, preserving its
dynamic type.

An unavailable or incompatible capability returns `present == false` without
an error. This includes nil, cyclic pointer, and interface subject types that
cannot supply a surrogate. Invalid capability requests, incompatible storage,
and hook panics return errors. A successful zero or nil result still returns
`present == true`.

Every call constructs a fresh zero-valued surrogate and invokes independently;
it does not access a live subject or cache results. The caller decides whether
to request a capability and how to use its result. Qualities instead retain
their compiled reader in the schema and require presence when activated.

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
