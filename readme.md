Sail
====

## Lane policies

Lanes form ordered bands in a subject's declaration. Spots are unordered
alternatives within a lane. The `exploration` tag declares how population
should choose between lanes that accept the current field:

| Policy | Meaning |
| --- | --- |
| `eager` | Prefer the current lane while it accepts the field. |
| `curious` | Prefer a reachable subsequent lane that also accepts the field. |
| `unambiguous` | Require the declaration to determine placement without a preference. |

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
