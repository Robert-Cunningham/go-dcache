package dcache2

type Operation byte

const (
	Create Operation = iota
	Rename
	Delete
)

/*
type Command struct {
	operation Operation
	path      Path
}
*/

type JobQueue struct {
	jobs []*Job
}

func mustWaitOn(later, earlier Command) bool {
	for _, trav := range later.getTraversals() {
		for _, targ := range earlier.getTargets() {
			if trav == targ {
				return true
			}
		}
	}
	return false
}
