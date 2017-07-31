package dcache2

import (
)

type hlist_bl_head struct {
	first *hlist_bl_node
	lock SpinLock
}

type hlist_bl_node struct {
	next *hlist_bl_node
	pprev **hlist_bl_node
	data interface{}
}

func __hlist_bl_del(b *hlist_bl_node) {
	prev := b.pprev
	next := b.next

	*prev = next

}

func (h *hlist_bl_head) hlist_bl_add_head(n *hlist_bl_node) { //HEAD -> OLD => HEAD -> NEW -> OLD
	first := h.first

	n.next = first
	if first != nil {
		first.pprev = &n.next
	}
	n.pprev = &h.first
	h.first = n
}

func (b *hlist_bl_node) init_hlist_bl_node() {
	b = new(hlist_bl_node)
}

func (h *hlist_bl_head) init_hlist_bl_head() {
	h = new(hlist_bl_head)
}

func (h *hlist_bl_node) hlist_bl_unhashed() bool { //is this node contained within a list?
	return h.pprev==nil
}

func (h *hlist_bl_head) hlist_bl_lock() {
	h.lock.Lock()
}

func (h *hlist_bl_head) hlist_bl_unlock() {
	h.lock.Unlock()
}

func (h *hlist_bl_head) hlist_bl_is_locked() bool {
	return h.lock.IsLocked()
}

func BL_BUG (test bool) {
	if test {
		panic("Bug in list_bl.go")
	}
}