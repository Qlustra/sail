Sail
====

## Population

`Explore[C]()` compiles the chart declaration once. `Populate[S](world)` then
returns a fresh `*C` populated against **S's type**, without creating an S
instance. The world contains reusable schema definitions; populated markers
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
