package dcache2

import "strings"

type Command interface {
	getTraversals() []string
	getTargets() []string
	execute()
}

// Traversals are things which cannot be modified while a job is in progress.
// Targets are things which might be modified by a job.

func getNonEmptySubStrings(in []string) []string {
	out := make([]string, len(in), len(in))
	current := ""
	for i := 0; i < len(in); i++ {
		current = current + "/" + in[i]
		out[i] = current
	}
	return out
}

type CreateOp struct {
	path Path
}

func (c CreateOp) execute() {

}

func (c CreateOp) getTraversals() []string {
	split := getNonEmptySubStrings(strings.Split(c.path.string[1:], "/"))
	out := split[0 : len(split)-1]
	return out
}

func (c CreateOp) getTargets() []string {
	out := make([]string, 1, 1)
	split := getNonEmptySubStrings(strings.Split(c.path.string[1:], "/"))
	out[0] = split[len(split)-1]
	return out
}

type DeleteOp struct {
	path Path
}

func (c DeleteOp) execute() {

}
func (c DeleteOp) getTraversals() []string {
	split := getNonEmptySubStrings(strings.Split(c.path.string[1:], "/"))
	out := split[0 : len(split)-1]
	return out
}

func (c DeleteOp) getTargets() []string {
	out := make([]string, 1, 1)
	split := getNonEmptySubStrings(strings.Split(c.path.string[1:], "/"))
	out[0] = split[len(split)-1]
	return out
}

type RenameOp struct {
	path    Path
	newName string
}

func (c RenameOp) execute() {

}

func (c RenameOp) getTraversals() []string {
	split := getNonEmptySubStrings(strings.Split(c.path.string[1:], "/"))
	out := split[0 : len(split)-1]
	return out
}

func (c RenameOp) getTargets() []string {
	out := make([]string, 1, 1)
	split := getNonEmptySubStrings(strings.Split(c.path.string[1:], "/"))
	out[0] = split[len(split)-1]
	return out
}
