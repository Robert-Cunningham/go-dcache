package dcache2

import "sync"

type Executor struct {
	staging []*Job
}

func (e *Executor) runJob(j *Job) {
	j.command.execute()
}

func (e *Executor) addCommand(c *Command) {

}

func (e *Executor) done(j *Job) {
	j.completed = true
	for _, d := range j.toStartOnMyFinish {
		d.waitingOn--
		e.checkReadyToExecute(d)
	}
}

func (e *Executor) NewJob(c Command) *Job {
	j := &Job{}
	j.command = c
	j.onComplete = func(e *Executor, j *Job) func() {
		return func() {
			e.done(j)
		}
	}(e, j)

	dependencies := j.determineDependencies(e.staging)
	j.addHooksToDependencies(dependencies)
	return j
}

type Job struct {
	command           Command
	onComplete        func()
	completed         bool
	waitingOn         int
	dependencyLock    sync.Mutex
	toStartOnMyFinish []*Job
}

func (e *Executor) checkReadyToExecute(j *Job) {
	if j.waitingOn != 0 {
		return
	}
	e.runJob(j)
}

func (j *Job) addHooksToDependencies(dependencies []*Job) {
	for _, d := range dependencies {
		d.dependencyLock.Lock()
		if !d.completed {
			j.waitingOn++
			d.toStartOnMyFinish = append(d.toStartOnMyFinish, j)
		}
		d.dependencyLock.Unlock()
	}
}

func (j *Job) determineDependencies(possibleConflicts []*Job) []*Job { // An earlier job can never wait on a later job.
	out := make([]*Job, 0, 0)
	for _, currentJob := range possibleConflicts {
		if mustWaitOn(j.command, currentJob.command) {
			out = append(out, currentJob)
		}
	}
	return out
}
