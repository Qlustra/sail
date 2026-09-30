package sail

import (
	"fmt"
	"reflect"
	"slices"
)

// ======================================== Journal

type locationIdentity reflect.Type

type Journal struct {
	visited map[locationIdentity]explorationState
	trace   []locationIdentity
}

func NewJournal() Journal {
	return Journal{
		visited: make(map[locationIdentity]explorationState),
	}
}

func (j *Journal) Visit(t locationIdentity) {
	if j.Seen(t) {
		return
	}
	if j.visited == nil {
		j.visited = make(map[locationIdentity]explorationState)
	}
	j.visited[t] = charting
	j.trace = append(j.trace, t)
}

func (j *Journal) Finish(t locationIdentity) error {
	if !j.Recorded(t) {
		return fmt.Errorf("invalid journal state: cannot finish charting: location not visited yet")
	}
	if j.Charted(t) {
		return nil
	}
	if len(j.trace) == 0 || j.trace[len(j.trace)-1] != t {
		return fmt.Errorf("invalid journal state: cannot finish charting %v before its nested locations", t)
	}
	j.visited[t] = charted
	j.trace = j.trace[:len(j.trace)-1]
	return nil
}

func (j *Journal) Traced(t locationIdentity) bool {
	return slices.Contains(j.trace, t)
}

func (j *Journal) Recorded(t locationIdentity) bool {
	_, exists := j.visited[t]
	return exists
}

func (j *Journal) Visiting(t locationIdentity) bool {
	state, exists := j.visited[t]
	return exists && state == charting
}

func (j *Journal) Seen(t locationIdentity) bool {
	state, exists := j.visited[t]
	return exists && (state == charting || state == charted)
}

func (j *Journal) Charted(t locationIdentity) bool {
	state, exists := j.visited[t]
	return exists && state == charted
}

type explorationState int

const (
	charting = iota
	charted
)
