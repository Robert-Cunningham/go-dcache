package dcache2

import (
)
import "fmt"

type list_node struct {
	next *list_node
	prev *list_node
	data interface{}
}

func list_add(toAdd *list_node, head *list_node) {
	if head.next == nil {
		head.next = &list_node{}
	}
	__list_add(toAdd, head, head.next)
}

func __list_add(toAdd *list_node, prev *list_node, next *list_node) {
	next.prev = toAdd
	toAdd.next = next
	toAdd.prev = prev
	prev.next = toAdd
}

func stringifyList(a *list_node) string {
	out := "LIST: {"

	for {
		out = out + fmt.Sprintf(", %+v ", a.data)
		if(a.next == nil) {
			break;
		} else {
			a = a.next
		}
	}


	out = out + "} "
	return out
}

func sliceifyList(a *list_node) []interface{} {
	var out []interface{} = make([]interface{}, 10)

	for {
		out = append(out, a.data)
		if(a.next == nil) {
			break;
		} else {
			a = a.next
		}
	}

	return out
}