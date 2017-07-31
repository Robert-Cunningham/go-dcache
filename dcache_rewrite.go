package dcache2

import (
	"fmt"
	"runtime"
	"sync/atomic"
)

type dcache_rw struct {

	dentry_hashtable [DENTRY_HASHTABLE_SIZE]hlist_bl_head
	//in_lookup_hashtable [1 << IN_LOOKUP_SHIFT]hlist_bl_head

	global_super_block super_block

	nr_dentry uint64
	nr_dentry_unused uint64
	rename_lock SeqLock
}

func (c *dcache_rw) getHomeListOfDentry(hash uint32) (*hlist_bl_head) {
	return &c.dentry_hashtable[hash % DENTRY_HASHTABLE_SIZE]
}

/*
func (c *dcache_rw) in_lookup_hash(parent *Dentry, hash uint32) (hlist_bl_head) {
	hash += uint32(getDentryIntUID(parent) / L1_CACHE_BYTES)
	return c.in_lookup_hashtable[hash_32(hash, IN_LOOKUP_SHIFT)]
}
*/

func dcache_rw_init() (*dcache_rw){

	c := dcache_rw{}

	//c.in_lookup_hashtable = [1 << IN_LOOKUP_SHIFT]hlist_bl_head{}

	c.global_super_block = super_block{}
	c.global_super_block.grandparent = c.__d_alloc(&c.global_super_block, "GLOBAL GRANDPARENT")

	for loop := 0; loop < DENTRY_HASHTABLE_SIZE; loop++ {
		c.dentry_hashtable[loop].init_hlist_bl_head()
	}

	pp.setDcache(&c)

	return &c
}

func (c *dcache_rw) getSuperBlock() (*super_block){
	return &c.global_super_block
}

func (c *dcache_rw) d_is_miss(dentry *Dentry) bool {
	return dentry.d_type == DENTRY_MISS_TYPE
}

func (c *dcache_rw) d_in_lookup(dentry *Dentry) bool {
	return (dentry.d_flags & DCACHE_PAR_LOOKUP != 0)
}

func (c *dcache_rw) d_add (dentry *Dentry, inode *Inode) {
	if inode != nil {
		inode.i_lock.Lock()
	}

	c.__d_add(dentry, inode)
}

func (c *dcache_rw) __d_add(dentry *Dentry, inode *Inode) {
	dentry.d_lock.Lock()

	if c.d_in_lookup(dentry) {
		//deal with what happens if a lookup is in progress. 1661. TODO
	}

	if inode != nil {
		dentry.d_seq.raw_write_seqcount_begin()
		dentry.d_type = c.d_flags_for_inode(inode)
		c.__d_set_inode_and_type(dentry, inode)
		dentry.d_seq.raw_write_seqcount_end()
	}

	c.__d_rehash(dentry)

	//other omitted stuff

	dentry.d_lock.Unlock()

	if inode != nil {
		inode.i_lock.Unlock()
	}

}

func (c *dcache_rw) __d_rehash(dentry *Dentry) {
	b := c.getHomeListOfDentry(getHashOfDentry(dentry)) //must already be dehashed at this point. Gets the home list for the dentry.
	//fmt.Printf("Insertion Hash: %+v \n", getHashOfDentry(dentry))
	if !c.d_unhashed(dentry) {
		fmt.Println("DISIASTER __d_rehash")
	}

	b.hlist_bl_lock()
	b.hlist_bl_add_head(&dentry.d_masterListNode) //TODO: 1486, RCU in original
	b.hlist_bl_unlock()

	//fmt.Printf("dentry_hashtable: %+v \n", dentry_hashtable)
}

func (c *dcache_rw) d_rehash(dentry *Dentry) {
	dentry.d_lock.Lock()
	c.__d_rehash(dentry)
	dentry.d_lock.Unlock()
}

func (c *dcache_rw) d_flags_for_inode(inode *Inode) dentryType {
	var dt dentryType = DENTRY_REGULAR_TYPE
	if inode == nil {
		return DENTRY_MISS_TYPE
	}

	if inode.i_mode==INODE_DIRECTORY_TYPE {
		dt = DENTRY_DIRECTORY_TYPE
	}

	return dt
}

func (c *dcache_rw) __d_set_inode_and_type(dentry *Dentry, inode *Inode) {
	dentry.d_inode = inode
	//dentry.d_flags = dentry.d_flags & ^(DCACHE_ENTRY_TYPE | DCACHE_FALLTHRU)
	dentry.d_type = DENTRY_REGULAR_TYPE //TODO: unclear
}

func (c *dcache_rw) __d_clear_type_and_inode(dentry *Dentry) {
	//dentry.d_flags = dentry.d_flags & ^(DCACHE_ENTRY_TYPE | DCACHE_FALLTHRU)
	dentry.d_inode = nil
	dentry.d_type = DENTRY_REGULAR_TYPE //TODO: Unclear
}

func (c *dcache_rw) dentry_unlink_inode(dentry *Dentry) {
	inode := dentry.d_inode
	hashed := !c.d_unhashed(dentry)

	if hashed {
		dentry.d_seq.raw_write_seqcount_begin()//write_seqbegin()
	}

	c.__d_clear_type_and_inode(dentry)

	if hashed {
		dentry.d_seq.raw_write_seqcount_end()
	}

	dentry.d_lock.Unlock()
	inode.i_lock.Unlock()


	c.iput(inode)
}

func (c *dcache_rw) d_delete(dentry *Dentry) {
	for {
		dentry.d_lock.Lock()
		inode := dentry.d_inode

		if(inode != nil && inode.i_lock.IsLocked()) {
			fmt.Println("Waiting for inode which is in use.")
			dentry.d_lock.Unlock()
			runtime.Gosched()
			continue
		} else {
			break;
		}
	}

	if(dentry.d_inode != nil) {
		c.dentry_unlink_inode(dentry)
	}

	if(!c.d_unhashed(dentry)) {
		c.__removeFromHashLists(dentry)
	}

	dentry.d_type = DENTRY_MISS_TYPE

	dentry.d_lock.Unlock()
}

func (c *dcache_rw) iput(inode *Inode) {
	inode.isDeleted = true
}

func (c *dcache_rw) __d_alloc(sb *super_block, name string) (*Dentry) {
	var dentry *Dentry = &Dentry{}

	if(&name == nil) {
		name = "/"
	} //we choose not to worry about whether name is too long (> DNAME_INLINE_LEN).

	dentry.d_name = name
	dentry.d_flags = 0
	dentry.d_lock = LockRef{}
	dentry.d_seq = SeqLock{}
	dentry.d_inode = nil
	dentry.d_parent = dentry //changed in parent function.
	dentry.d_sb = sb
	dentry.d_masterListNode = hlist_bl_node{data:dentry}
	dentry.d_lru = list_node{}
	dentry.d_subdirs = list_node{} // data: dentry} ??

	atomic.AddUint64(&c.nr_dentry, 1)

	return dentry
}

func (c *dcache_rw) d_alloc(parent *Dentry, name string) (*Dentry) {
	dentry := c.__d_alloc(parent.d_sb, name)
	if &dentry == nil {
		return nil
	}
	dentry.d_flags = dentry.d_flags | DCACHE_RCUACCESS
	parent.d_lock.Lock()
	parent.__dget_dlock()

	dentry.d_parent = parent
	list_add(&list_node{data:dentry}, &parent.d_subdirs)

	parent.d_lock.Unlock()

	return dentry
}

func (dentry *Dentry) __dget_dlock() { //don't understand why this function is useful.
	dentry.d_lock.count++
}

func (c *dcache_rw) __d_lookup_ref(parent *Dentry, name string) *Dentry {
	var toFindHash uint32 = getHashOfParentAndName(parent, name)
	//fmt.Printf("Retreival Hash: %+v", toFindHash)
	//fmt.Printf("dentry_hashtable: %+v \n", dentry_hashtable)
	var head *hlist_bl_head = c.getHomeListOfDentry(toFindHash) //the head with the correct hash.
	var node *hlist_bl_node
	var found *Dentry
	var dentry *Dentry

	node = head.first

	if node == nil {
		return nil
	}

	dentry = node.data.(*Dentry)

	for {
		dentry.d_lock.Lock()

		if getHashOfDentry(dentry) != toFindHash {
			//fmt.Printf("Dentry hash Unequal \n")
			goto next
		}

		if dentry.d_parent != parent {
			//fmt.Printf("Parent Unequal \n")
			goto next
		}

		if c.d_unhashed(dentry) {
			//fmt.Printf("Rip3 \n")
			goto next
		}

		if !(dentry.d_name == name) {
			//fmt.Printf("Name unequal \n")
			goto next
		}

		//dentry.d_lock.count++
		found = dentry
		dentry.d_lock.Unlock()
		//fmt.Printf("__d_lookup out: %+v \n", found)
		break

		next: //failed
		dentry.d_lock.Unlock()

		node = node.next
		if node == nil {
			return nil
		}
		dentry = node.data.(*Dentry)
	}

	return found
}


func (c *dcache_rw) __d_lookup_rcu(parent *Dentry, name string) *Dentry {
	var toFindHash uint32 = getHashOfParentAndName(parent, name)
	var head *hlist_bl_head = c.getHomeListOfDentry(toFindHash) //the head with the correct hash.
	var node *hlist_bl_node
	var found *Dentry
	var dentry *Dentry

	node = head.first

	if node == nil {
		return nil
	}

	dentry = node.data.(*Dentry)

	for {
		seq := dentry.d_seq.read_seqbegin()

		if getHashOfDentry(dentry) != toFindHash {
			//fmt.Printf("Dentry hash Unequal \n")
			goto next
		}

		if dentry.d_parent != parent {
			//fmt.Printf("Parent Unequal \n")
			goto next
		}

		if c.d_unhashed(dentry) {
			//fmt.Printf("Rip3 \n")
			goto next
		}

		if !(dentry.d_name == name) {
			//fmt.Printf("Name unequal \n")
			goto next
		}

		found = dentry

		if(dentry.d_seq.read_seqretry(seq)) {
			continue
		} else {
			break
		}

		next: //failed

		node = node.next
		if node == nil {
			return nil
		}
		dentry = node.data.(*Dentry)
	}

	return found
}

func (c *dcache_rw) d_lookup(parent *Dentry, name string) *Dentry {
	var seq uint32
	var dentry *Dentry
	for {
		seq = c.rename_lock.read_seqbegin()
		dentry = c.__d_lookup_ref(parent, name)
		if dentry != nil || !c.rename_lock.read_seqretry(seq) {
			break;
		}
	}

	return dentry
}

func (c *dcache_rw) is_root(dentry *Dentry) bool{
	return dentry == dentry.d_parent
}

func (c *dcache_rw) __removeFromHashLists(dentry *Dentry) { //requires d.d_lock
	if ! c.d_unhashed(dentry) {
		var b *hlist_bl_head

		if(c.is_root(dentry)) {
			b = &dentry.d_sb.s_anon //might get rid of this
		} else {
			b = c.getHomeListOfDentry(getHashOfDentry(dentry))
		}

		b.hlist_bl_lock()
		__hlist_bl_del(&dentry.d_masterListNode)
		dentry.d_masterListNode.pprev = nil
		b.hlist_bl_unlock()

		dentry.d_seq.write_seqlock_invalidate()
	}
}

func (c *dcache_rw) removeFromHashLists(dentry *Dentry) {
	dentry.d_lock.Lock()
	c.__removeFromHashLists(dentry)
	dentry.d_lock.Unlock()
}

func (c *dcache_rw) d_unhashed(dentry *Dentry) bool {
	return dentry.d_masterListNode.hlist_bl_unhashed()
}

func (c *dcache_rw) d_is_negative(dentry *Dentry) bool {
	return dentry.d_type == DENTRY_MISS_TYPE
}