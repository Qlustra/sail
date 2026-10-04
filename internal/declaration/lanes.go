package declaration

type LaneExplorationStrategy string

const (
	ExploreEagerly       LaneExplorationStrategy = "eager"
	ExploreUnambiguously LaneExplorationStrategy = "unambiguous"
)

type Lane struct{}

func (Lane) lane() {}

type LaneDeclaration interface {
	lane()
}
