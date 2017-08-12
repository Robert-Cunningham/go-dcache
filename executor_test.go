package dcache2

import (
	"fmt"
	"testing"
	"time"
)

func TestGetTargets(t *testing.T) {
	assertGetTargets(t, CreateOp{Path{"/a/b/c/d"}}, []string{"/a/b/c/d"})
	assertGetTargets(t, CreateOp{Path{"/a"}}, []string{"/a"})
}

func TestGetTraversals(t *testing.T) {
	assertGetTraversals(t, CreateOp{Path{"/a/b/c/d"}}, []string{"/a", "/a/b", "/a/b/c"})
	assertGetTraversals(t, CreateOp{Path{"/a"}}, []string{})
}

func TestMustWaitOn(t *testing.T) {
	assertMustWaitOn(t, Path{"/a/b/c"}, Path{"/a/b"}, true)
	assertMustWaitOn(t, Path{"/a/b/c"}, Path{"/a/b/d"}, false)
	assertMustWaitOn(t, Path{"/a/b/d"}, Path{"/a"}, true)
	assertMustWaitOn(t, Path{"/a"}, Path{"/b/f/d"}, false)
	assertMustWaitOn(t, Path{"/a/b/c/d"}, Path{"/a/b/c/d/e"}, true)
}

func assertMustWaitOn(t *testing.T, p1 Path, p2 Path, expected bool) {
	c1 := CreateOp{p1}
	c2 := CreateOp{p2}
	if mustWaitOn(c1, c2) != expected {
		fmt.Printf("Failed %v and %v.", p1, p2)
		t.Fail()
	}
}

func assertGetTargets(t *testing.T, c Command, expectedResult []string) {
	out := c.getTargets()
	if len(out) != len(expectedResult) {
		fmt.Printf("Got %v, but expected %v. ", out, expectedResult)
		t.Fail()
		return
	}
	for i := range expectedResult {
		if expectedResult[i] != out[i] {
			fmt.Printf("Got %v, but expected %v. ", out, expectedResult)
			t.Fail()
		}
	}
}

func assertGetTraversals(t *testing.T, c Command, expectedResult []string) {
	out := c.getTraversals()
	if len(out) != len(expectedResult) {
		fmt.Printf("Got %v, but expected %v. ", out, expectedResult)
		t.Fail()
		return
	}
	for i := range expectedResult {
		if expectedResult[i] != out[i] {
			fmt.Printf("Got %v, but expected %v. ", out, expectedResult)
			t.Fail()
		}
	}
}

func TestAssertOrdering(t *testing.T) {
	e := NewExecutor(2)
	order := make([]int, 0, 0)
	onFinish := func(id int) func() {
		return func() {
			fmt.Println(id, "started")
			time.Sleep(10 * time.Millisecond)
			order = append(order, id)
			fmt.Println(id, "ended")
		}
	}
	e.addCommand(TestOp{onFinish(2), CreateOp{Path{"/a/b/c/d/e"}}}, 1)
	time.Sleep(time.Millisecond)
	e.addCommand(TestOp{onFinish(1), CreateOp{Path{"/a/b/c/d"}}}, 2)
	time.Sleep(time.Millisecond)
	e.addCommand(TestOp{onFinish(3), CreateOp{Path{"/a/b/c/d/f"}}}, 3)
	time.Sleep(time.Millisecond)
	e.addCommand(TestOp{onFinish(5), CreateOp{Path{"/a/g/h/"}}}, 4) //This one should finish before the other two.
	time.Sleep(time.Millisecond)
	e.addCommand(TestOp{onFinish(4), CreateOp{Path{"/a/b/c"}}}, 5)
	time.Sleep(time.Millisecond)

	time.Sleep(1000 * time.Millisecond)

	fmt.Println(order)

	//order[0] == 1 && order[1] == 4 && order
}
