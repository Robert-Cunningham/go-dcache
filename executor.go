package dcache2

import (
	"fmt"
	"sort"
	"sync"
	"sync/atomic"
)

type Executor struct {
	waiting           []*Job
	staging           []*Job
	waitingLock       sync.Mutex
	stagingLock       sync.Mutex
	maxConcurrentOps  int32
	currentOpsRunning int32
}

func NewExecutor(concurrentOps int32) *Executor {
	return &Executor{make([]*Job, 0, 0), make([]*Job, 0, 0), sync.Mutex{}, sync.Mutex{}, concurrentOps, 0}
}

func (e *Executor) runJob(j *Job) {
	go func() {
		atomic.AddInt32(&e.currentOpsRunning, 1)
		j.running = true
		j.command.execute()
		j.onComplete()
		e.cleanUp(j)
		atomic.AddInt32(&e.currentOpsRunning, -1)
	}()
}

func (e *Executor) addToWaiting(j *Job) {
	e.waitingLock.Lock()
	e.waiting = append(e.waiting, j)
	e.waitingLock.Unlock()
}

func (e *Executor) addToStaging(j *Job) {
	e.stagingLock.Lock()
	e.staging = append(e.staging, j)
	e.stagingLock.Unlock()
}

func removeFromList(list *([]*Job), lock *sync.Mutex, j *Job) {
	for i, s := range *list {
		if s == j {
			lock.Lock()
			out := append((*list)[:i], (*list)[i+1:]...)
			list = &out
			lock.Unlock()
			break
		}
	}
}

func (e *Executor) addCommand(c Command, priority int) {
	j := e.NewJob(c, priority)
	e.addToWaiting(j)
	e.tryToMoveFromWaitingToStaging(j)
}

func (e *Executor) tryToMoveFromWaitingToStaging(j *Job) {
	if e.checkReadyToExecute(j) {
		e.pushToStaging(j)
		e.tryToPopFromStaging()
	}
}

func (e *Executor) cleanUp(j *Job) {
	j.completed = true

	for _, d := range j.toStartOnMyFinish {
		fmt.Printf("%v left for %v:%v. \n ", d.waitingOn, d.command.getTraversals(), d.command.getTargets())
		d.waitingOn--
		e.tryToMoveFromWaitingToStaging(d)
	}
}

func (e *Executor) pushToStaging(j *Job) {
	e.addToStaging(j)
}

func (e *Executor) popFromStaging() bool {
	if len(e.staging) == 0 {
		return false
	}
	e.stagingLock.Lock()
	sort.Slice(e.staging, func(i, j int) bool {
		return e.staging[i].priority < e.staging[j].priority
	})
	e.runJob(e.staging[len(e.staging)-1])
	e.staging = e.staging[:len(e.staging)-1]
	e.stagingLock.Unlock()
	return true
}

func (e *Executor) NewJob(c Command, priority int) *Job {
	j := &Job{}
	j.command = c
	j.onComplete = func(e *Executor, j *Job) func() {
		return func() {
		}
	}(e, j)

	fmt.Printf("waiting:%v, staging:%v \n", e.waiting, e.staging)

	dependencies := j.determineDependencies(e.waiting)
	dependencies2 := j.determineDependencies(e.staging)
	sum := append(dependencies, dependencies2...)
	fmt.Printf("%v, just added, must wait on %v. \n", c, sum)
	j.addHooksToDependencies(sum)
	return j
}

func (e *Executor) tryToPopFromStaging() {
	if atomic.LoadInt32(&e.currentOpsRunning) < e.maxConcurrentOps {
		success := e.popFromStaging()
		if success {
			e.tryToPopFromStaging()
		}
	}
}

type Job struct {
	command           Command
	onComplete        func()
	completed         bool
	running           bool
	waitingOn         int
	dependencyLock    sync.Mutex
	toStartOnMyFinish []*Job
	priority          int
}

func (e *Executor) checkReadyToExecute(j *Job) bool {
	if j.waitingOn != 0 {
		return false
	}
	return true
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
