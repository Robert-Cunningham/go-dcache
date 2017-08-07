package dcache2

import (
	"strings"
)

type Operation byte

const (
	Create Operation = iota
	Rename
	Delete
)

type Job struct {
	operation Operation
	path      Path
	waitOn    []*Job
	completed bool
}

type JobQueue struct {
	jobs []*Job
}

func NewJob(p Path, op Operation) *Job {
	return &Job{
		operation: op,
		path:      p,
		waitOn:    nil,
		completed: false,
	}

}

// Traversals are things which cannot be modified while a job is in progress.
func (j *Job) getTraversals() []string {
	split := strings.Split(j.path.original, "/")
	out := split[:len(split)-1]
	return out
}

// Targets are things which might be modified by a job.
func (j *Job) getTargets() []string {
	if j.operation == Create {
		return make([]string, 0, 0)
	}
	if j.operation == Rename || j.operation == Delete {
		out := make([]string, 1, 1)
		split := strings.Split(j.path.original, "/")
		out[0] = split[len(split)-1]
	}
	return nil
}

func (j *Job) mustWaitOn(earlier *Job) bool {
	for _, trav := range j.getTraversals() {
		for _, targ := range earlier.getTargets() {
			if trav == targ {
				return true
			}
		}
	}
	return false
}

func (j *Job) determineDependencies(possibleConflicts []*Job) { // An earlier job can never wait on a later job.
	for _, currentJob := range possibleConflicts {
		if j.mustWaitOn(currentJob) {
			j.waitOn = append(j.waitOn, currentJob)
		}
	}
}
