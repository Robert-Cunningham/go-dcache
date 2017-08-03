package dcache2

type lockingList_head struct {
	first *lockingList_node
	lock  SpinLock
}

type lockingList_node struct {
	next  *lockingList_node
	pprev **lockingList_node
	data  interface{}
}

func __lockingList_del(b *lockingList_node) {
	prev := b.pprev
	next := b.next

	*prev = next

}

func (h *lockingList_head) lockingList_add_head(n *lockingList_node) { //HEAD -> OLD => HEAD -> NEW -> OLD
	first := h.first

	n.next = first
	if first != nil {
		first.pprev = &n.next
	}
	n.pprev = &h.first
	h.first = n
}

func (b *lockingList_node) init_lockingList_node() {
	b = new(lockingList_node)
}

func (h *lockingList_head) init_lockingList_head() {
	h = new(lockingList_head)
}

func (h *lockingList_node) lockingList_unhashed() bool { //is this node contained within a list?
	return h.pprev == nil
}

func (h *lockingList_head) lockingList_lock() {
	h.lock.Lock()
}

func (h *lockingList_head) lockingList_unlock() {
	h.lock.Unlock()
}

func (h *lockingList_head) lockingList_is_locked() bool {
	return h.lock.IsLocked()
}

func BL_BUG(test bool) {
	if test {
		panic("Bug in list_bl.go")
	}
}
