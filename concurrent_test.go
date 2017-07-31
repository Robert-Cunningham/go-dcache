package dcache2

import (
	"testing"
)

func concurrentCreator(t *testing.T) {

}

func concurrentMover(t *testing.T) {

}

func concurrentDeleter(t *testing.T) {

}

func concurrentRenamer(t *testing.T) {

}

func concurrentModifier(t *testing.T) {

}

func TestConcurrency(t *testing.T) {
	go concurrentCreator(t)
	go concurrentMover(t)
	go concurrentModifier(t)
	//go concurrentDeleter(t)

}