package dcache2

import (
	"fmt"
	"math/rand"
	"runtime"
)

type Dcache struct {
	dentry_hashtable  [DENTRY_HASHTABLE_SIZE]lockingList_head
	global_SuperBlock SuperBlock
	rename_lock       SeqLock
}

var globalSuperBlock SuperBlock

func (c *Dcache) getHomeListOfDentry(hash uint32) *lockingList_head {
	return &c.dentry_hashtable[hash%DENTRY_HASHTABLE_SIZE]
}

func InitializeDcache() *Dcache {
	c := Dcache{}
	c.global_SuperBlock = SuperBlock{}
	c.global_SuperBlock.grandparent = c.__NewDentry(&c.global_SuperBlock, "GLOBAL GRANDPARENT")

	for loop := 0; loop < DENTRY_HASHTABLE_SIZE; loop++ {
		c.dentry_hashtable[loop].init_lockingList_head()
	}

	pp.setDcache(&c)

	return &c
}

func (c *Dcache) dput(d *Dentry) {
	//TODO. Need to remove the dentry from all the appropriate lists.
	c.d_delete(d)
}

func (c *Dcache) getSuperBlock() *SuperBlock {
	return &c.global_SuperBlock
}

func (c *Dcache) d_is_miss(dentry *Dentry) bool {
	return dentry.d_type == DENTRY_MISS_TYPE
}

func (c *Dcache) d_in_lookup(dentry *Dentry) bool {
	return (dentry.d_flags&DCACHE_PAR_LOOKUP != 0)
}

func (c *Dcache) d_add(dentry *Dentry, inode *Inode) {
	if inode != nil {
		inode.i_lock.Lock()
	}

	c.__d_add(dentry, inode)
}

func (c *Dcache) __d_add(dentry *Dentry, inode *Inode) {
	dentry.d_lock.Lock()

	if c.d_in_lookup(dentry) {
		//deal with what happens if a lookup is in progress. 1661. TODO
	}

	if inode != nil {
		dentry.d_seq.raw_write_seqcount_begin()
		dentry.setInode(inode)
		dentry.d_seq.raw_write_seqcount_end()
	}

	c.__d_rehash(dentry)

	//other omitted stuff

	dentry.d_lock.Unlock()

	if inode != nil {
		inode.i_lock.Unlock()
	}

}

func (c *Dcache) __d_rehash(dentry *Dentry) {
	b := c.getHomeListOfDentry(getHashOfDentry(dentry)) //must already be dehashed at this point. Gets the home list for the dentry.
	//fmt.Printf("Adding %v to list %v with hash %v. \n", dentry.d_name, b, getHashOfDentry(dentry))
	//fmt.Printf("Insertion Hash: %+v \n", getHashOfDentry(dentry))
	if !c.DentryNotInDcache(dentry) {
		fmt.Println("DISIASTER __d_rehash")
	}

	b.lockingList_lock()
	b.lockingList_add_head(&dentry.d_masterListNode) //TODO: 1486, RCU in original
	b.lockingList_unlock()

	//fmt.Printf("dentry_hashtable: %+v \n", dentry_hashtable)
}

func (c *Dcache) d_rehash(dentry *Dentry) {
	dentry.d_lock.Lock()
	c.__d_rehash(dentry)
	dentry.d_lock.Unlock()
}

func (d *Dentry) setInode(inode *Inode) {
	if inode == nil {
		d.d_type = DENTRY_MISS_TYPE
		d.d_inode = nil
		return
	}
	if inode.i_mode == INODE_DIRECTORY_TYPE {
		d.d_type = DENTRY_DIRECTORY_TYPE
		return
	}

	d.d_inode = inode
}

func (d *Dentry) clearInode() {
	d.setInode(nil)
}

func (c *Dcache) dentry_unlink_inode(dentry *Dentry) {
	inode := dentry.d_inode
	hashed := !c.DentryNotInDcache(dentry)

	if hashed {
		dentry.d_seq.raw_write_seqcount_begin() //write_seqbegin()
	}

	dentry.clearInode()

	if hashed {
		dentry.d_seq.raw_write_seqcount_end()
	}

	dentry.d_lock.Unlock()
	inode.i_lock.Unlock()

	c.DeleteInode(inode)
}

func (c *Dcache) d_delete(dentry *Dentry) {
	for {
		dentry.d_lock.Lock()
		inode := dentry.d_inode

		if inode != nil && inode.i_lock.IsLocked() {
			fmt.Println("Waiting for inode which is in use.")
			dentry.d_lock.Unlock()
			runtime.Gosched()
			continue
		} else {
			break
		}
	}

	if dentry.d_inode != nil {
		c.dentry_unlink_inode(dentry)
	}

	if !c.DentryNotInDcache(dentry) {
		c.__removeFromHashLists(dentry)
	}

	dentry.d_type = DENTRY_MISS_TYPE

	dentry.d_lock.Unlock()
}

func (c *Dcache) DeleteInode(inode *Inode) { //iput
	inode.isDeleted = true
}

func (c *Dcache) __NewDentry(sb *SuperBlock, name string) *Dentry { //__d_alloc
	var dentry *Dentry = &Dentry{}

	if &name == nil {
		name = "/"
	} //we choose not to worry about whether name is too long (> DNAME_INLINE_LEN).

	dentry.d_name = name
	dentry.d_flags = 0
	dentry.d_lock = LockRef{}
	dentry.d_seq = SeqLock{}
	dentry.d_inode = nil
	dentry.d_parent = dentry //changed in parent function.
	dentry.d_sb = sb
	dentry.d_masterListNode = lockingList_node{data: dentry}
	dentry.d_lru = list_node{}
	dentry.d_subdirs = list_node{} // data: dentry} ??
	dentry.d_uuid = rand.Int()

	return dentry
}

func (c *Dcache) NewDentry(parent *Dentry, name string) *Dentry {
	dentry := c.__NewDentry(parent.d_sb, name)
	if &dentry == nil {
		return nil
	}
	dentry.d_flags = dentry.d_flags | DCACHE_RCUACCESS
	parent.d_lock.Lock()
	parent.__dget_dlock()

	dentry.d_parent = parent
	list_add(&list_node{data: dentry}, &parent.d_subdirs)

	parent.d_lock.Unlock()

	return dentry
}

func (dentry *Dentry) __dget_dlock() { //don't understand why this function is useful.
	dentry.d_lock.count++
}

func (c *Dcache) __d_lookup_ref(parent *Dentry, name string) *Dentry {
	var toFindHash uint32 = getHashOfParentAndName(parent, name)
	//fmt.Printf("Retreival Hash: %+v", toFindHash)
	//fmt.Printf("dentry_hashtable: %+v \n", c.dentry_hashtable)
	var head *lockingList_head = c.getHomeListOfDentry(toFindHash) //the head with the correct hash.
	var node *lockingList_node
	var found *Dentry
	var dentry *Dentry

	node = head.first

	if node == nil {
		return nil
	}

	dentry = node.data.(*Dentry)

	for {
		dentry.d_lock.Lock()

		if getHashOfDentry(dentry) == toFindHash && dentry.d_parent == parent && !c.DentryNotInDcache(dentry) && dentry.d_name == name {
			found = dentry
			dentry.d_lock.Unlock()
			return found
		}

		dentry.d_lock.Unlock()

		node = node.next
		if node == nil {
			return nil
		}
		dentry = node.data.(*Dentry)
	}
}

func (c *Dcache) d_lookup(parent *Dentry, name string) *Dentry {
	var seq uint32
	var dentry *Dentry
	for {
		seq = c.rename_lock.read_seqbegin()
		dentry = c.__d_lookup_ref(parent, name)
		if dentry != nil || !c.rename_lock.read_seqretry(seq) {
			break
		}
	}

	return dentry
}

func (c *Dcache) is_root(dentry *Dentry) bool {
	return dentry == dentry.d_parent
}

func (c *Dcache) __removeFromHashLists(dentry *Dentry) { //requires d.d_lock
	if !c.DentryNotInDcache(dentry) {
		var b *lockingList_head

		if c.is_root(dentry) {
			panic("Can't remove root from hash lists!")
		} else {
			b = c.getHomeListOfDentry(getHashOfDentry(dentry))
		}

		b.lockingList_lock()
		__lockingList_del(&dentry.d_masterListNode)
		dentry.d_masterListNode.pprev = nil
		b.lockingList_unlock()

		dentry.d_seq.write_seqlock_invalidate()
	}
}

func (c *Dcache) removeFromHashLists(dentry *Dentry) {
	dentry.d_lock.Lock()
	c.__removeFromHashLists(dentry)
	dentry.d_lock.Unlock()
}

func (c *Dcache) DentryNotInDcache(dentry *Dentry) bool { //d_unhashed
	return dentry.d_masterListNode.lockingList_unhashed()
}

func (c *Dcache) d_is_negative(dentry *Dentry) bool {
	return dentry.d_type == DENTRY_MISS_TYPE
}
