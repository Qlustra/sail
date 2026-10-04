package engine

import (
	"fmt"
	"reflect"
	"slices"
)

// ======================================== Journal

type declarationIdentity reflect.Type

type Journal struct {
	visited map[declarationIdentity]explorationState
	trace   []declarationIdentity
}

func NewJournal() Journal {
	return Journal{
		visited: make(map[declarationIdentity]explorationState),
	}
}

func (j *Journal) Visit(t declarationIdentity) {
	if j.Seen(t) {
		return
	}
	if j.visited == nil {
		j.visited = make(map[declarationIdentity]explorationState)
	}
	j.visited[t] = charting
	j.trace = append(j.trace, t)
}

func (j *Journal) Finish(t declarationIdentity) error {
	if !j.Recorded(t) {
		return fmt.Errorf("invalid journal state: cannot finish charting: descriptor not visited yet")
	}
	if j.Charted(t) {
		return nil
	}
	if len(j.trace) == 0 || j.trace[len(j.trace)-1] != t {
		return fmt.Errorf("invalid journal state: cannot finish charting %v before its nested descriptors", t)
	}
	j.visited[t] = charted
	j.trace = j.trace[:len(j.trace)-1]
	return nil
}

func (j *Journal) Traced(t declarationIdentity) bool {
	return slices.Contains(j.trace, t)
}

func (j *Journal) Recorded(t declarationIdentity) bool {
	_, exists := j.visited[t]
	return exists
}

func (j *Journal) Visiting(t declarationIdentity) bool {
	state, exists := j.visited[t]
	return exists && state == charting
}

func (j *Journal) Seen(t declarationIdentity) bool {
	state, exists := j.visited[t]
	return exists && (state == charting || state == charted)
}

func (j *Journal) Charted(t declarationIdentity) bool {
	state, exists := j.visited[t]
	return exists && state == charted
}

type explorationState int

const (
	charting = iota
	charted
)
