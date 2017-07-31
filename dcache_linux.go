package dcache2

import (
	"unsafe"
	"fmt"
	"sync/atomic"
	"runtime"
)

/*
	container_of(pointer, struct, name) assumes that _pointer_ is a pointer to the object with _name_ in _struct_.
*/

type dcache_l struct {

	dentry_hashtable [DENTRY_HASHTABLE_SIZE]hlist_bl_head
	in_lookup_hashtable [1 << IN_LOOKUP_SHIFT]hlist_bl_head
	dentry_cache Kmem_cache

	global_super_block super_block

	//var d_hash_mask uint
	//var d_hash_shift uint

	nr_dentry uint64
	nr_dentry_unused uint64
	rename_lock SeqLock

}

func (d *dcache_l) d_hash(hash uint32) (*hlist_bl_head) {
	return d.getHomeListOfDentry(hash)
}

func (d *dcache_l) getHomeListOfDentry(hash uint32) (*hlist_bl_head) {
	//fmt.Printf("Final index: %v \n", hash >> (32 - d_hash_shift))
	//return dentry_hashtable[hash >> (32 - d_hash_shift)]
	return &d.dentry_hashtable[hash % DENTRY_HASHTABLE_SIZE]
}

func (d *dcache_l) in_lookup_hash(parent *Dentry, hash uint32) (hlist_bl_head) {
	hash += uint32(getDentryIntUID(parent) / L1_CACHE_BYTES)
	return d.in_lookup_hashtable[hash_32(hash, IN_LOOKUP_SHIFT)]
}

func (d *dcache_l) alloc_dentry_hashtable(bucketsize uint64, numentries uint64, scale int, flags int) ([DENTRY_HASHTABLE_SIZE]hlist_bl_head){ //alloc_large_system_hash equivalent
	//my_dentry_hashtable := make([]hlist_bl_head, DENTRY_HASHTABLE_SIZE)
	return d.dentry_hashtable
}

func (d *dcache_l) dcache_init() {

	d.in_lookup_hashtable = [1 << IN_LOOKUP_SHIFT]hlist_bl_head{}

	d.global_super_block = super_block{}
	d.global_super_block.grandparent = d.__d_alloc(&d.global_super_block, "GLOBAL GRANDPARENT")

	d.dentry_cache = New_kmem_cache()

	d.dentry_hashtable = d.alloc_dentry_hashtable(uint64(unsafe.Sizeof(hlist_bl_head{})), dhash_entries, 13, 0)

	for loop := 0; loop < DENTRY_HASHTABLE_SIZE; loop++ {
		d.dentry_hashtable[loop].init_hlist_bl_head()
	}
}

func (d *dcache_l) d_backing_inode(dentry *Dentry) (*Inode) {
	return dentry.d_inode
}

func (d *dcache_l) d_is_miss(dentry *Dentry) bool {
	return dentry.d_type == DENTRY_MISS_TYPE
}

func (d *dcache_l) d_in_lookup(dentry *Dentry) bool {
	return (dentry.d_flags & DCACHE_PAR_LOOKUP != 0)
}

func (d *dcache_l) d_add (dentry *Dentry, inode *Inode) {
	if inode != nil {
		inode.i_lock.Lock()
	}

	d.__d_add(dentry, inode)
}

func (d *dcache_l) __d_add(dentry *Dentry, inode *Inode) {
	dentry.d_lock.Lock()

	if d.d_in_lookup(dentry) {
		//deal with what happens if a lookup is in progress. 1661. TODO
	}

	if inode != nil {
		dentry.d_seq.raw_write_seqcount_begin()
		dentry.d_type = d.d_flags_for_inode(inode)
		d.__d_set_inode_and_type(dentry, inode)
		dentry.d_seq.raw_write_seqcount_end()
	}

	d.__d_rehash(dentry)

	//other omitted stuff

	dentry.d_lock.Unlock()

	if inode != nil {
		inode.i_lock.Unlock()
	}

}


func (d *dcache_l) __d_rehash(dentry *Dentry) {
	b := d.getHomeListOfDentry(getHashOfDentry(dentry)) //must already be dehashed at this point. Gets the home list for the dentry.
	//fmt.Printf("Insertion Hash: %+v \n", getHashOfDentry(dentry))
	if !d.d_unhashed(dentry) {
		fmt.Println("DISIASTER __d_rehash")
	}

	b.hlist_bl_lock()
	b.hlist_bl_add_head(&dentry.d_masterListNode) //TODO: 1486, RCU in original
	b.hlist_bl_unlock()

	//fmt.Printf("dentry_hashtable: %+v \n", dentry_hashtable)
}

func (d *dcache_l) d_rehash(dentry *Dentry) {
	dentry.d_lock.Lock()
	d.__d_rehash(dentry)
	dentry.d_lock.Unlock()
}

func (d *dcache_l) d_flags_for_inode(inode *Inode) dentryType {
	var dt dentryType = DENTRY_REGULAR_TYPE
	if inode == nil {
		return DENTRY_MISS_TYPE
	}

	if inode.i_mode==INODE_DIRECTORY_TYPE {
		dt = DENTRY_DIRECTORY_TYPE
	}

	return dt
}

func (d *dcache_l) __d_set_inode_and_type(dentry *Dentry, inode *Inode) {
	dentry.d_inode = inode
	//dentry.d_flags = dentry.d_flags & ^(DCACHE_ENTRY_TYPE | DCACHE_FALLTHRU)
	dentry.d_type = DENTRY_REGULAR_TYPE //TODO: unclear
}

func (d *dcache_l) __d_clear_type_and_inode(dentry *Dentry) {
	//dentry.d_flags = dentry.d_flags & ^(DCACHE_ENTRY_TYPE | DCACHE_FALLTHRU)
	dentry.d_inode = nil
	dentry.d_type = DENTRY_REGULAR_TYPE //TODO: Unclear
}

func (d *dcache_l) dentry_unlink_inode(dentry *Dentry) {
	inode := dentry.d_inode
	hashed := !d.d_unhashed(dentry)

	if hashed {
		dentry.d_seq.raw_write_seqcount_begin()//write_seqbegin()
	}

	d.__d_clear_type_and_inode(dentry)

	if hashed {
		dentry.d_seq.raw_write_seqcount_end()
	}

	dentry.d_lock.Unlock()
	inode.i_lock.Unlock()


	d.iput(inode)
}

func (d *dcache_l) d_delete(dentry *Dentry) {
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
		d.dentry_unlink_inode(dentry)
	}

	if(!d.d_unhashed(dentry)) {
		d.__d_drop(dentry)
	}

	dentry.d_type = DENTRY_MISS_TYPE

	dentry.d_lock.Unlock()
}

func (d *dcache_l) iput(inode *Inode) {
	inode.isDeleted = true
}

func (d *dcache_l) dentry_cmp(dentry *Dentry, toCompare string) bool {
	//return strings.Compare(dentry.d_name.name, toCompare)
	return (dentry.d_name == toCompare)
}

func (d *dcache_l) dget_parent(child Dentry) Dentry {
	fmt.Println("dget_parent()")
	return *new(Dentry)
}

func (d *dcache_l) __d_alloc(sb *super_block, name string) (*Dentry) {
	var dentry *Dentry = &Dentry{}
	//var dname string
	//var err int

	d.dentry_cache.Kmem_cache_insert(dentry)

	if(&name == nil) {
		name = "/"
	} //we choose not to worry about whether name is too long (> DNAME_INLINE_LEN).

	//dname = dentry.d_iname

	//dentry.d_name.name = name.name
	//dentry.d_name.string_len = name.string_len
	//dentry.d_name.hash = name.hash

	/// memcpy(dname, name->name, name->len); existed here originally. Not sure of it's purpose.

	dentry.d_name = name
	//dentry.lockref.count = 1 not necessary due to GC
	dentry.d_flags = 0
	dentry.d_lock = LockRef{}
	dentry.d_seq = SeqLock{}
	dentry.d_inode = nil
	dentry.d_parent = dentry //changed in parent function.
	dentry.d_sb = sb
	//dentry.d_op = nil
	dentry.d_masterListNode = hlist_bl_node{data:dentry}
	dentry.d_lru = list_node{}
	dentry.d_subdirs = list_node{} // data: dentry} ??
	//dentry.d_sisters = list_node{} // data: dentry} ??
	//dentry.d_alias = hlist_bl_node{}

	d.d_set_d_op(dentry, &dentry.d_sb.s_d_op)

	/*
	if &dentry.d_op.d_init != nil {
		err = dentry.d_op.d_init(dentry)

		if err != 0{
			dentry_cache.Kmem_cache_free(dentry)
			return nil
		}
	}
	*/

	atomic.AddUint64(&d.nr_dentry, 1)

	return dentry
}

func (d *dcache_l) d_alloc(parent *Dentry, name string) (*Dentry) {
	dentry := d.__d_alloc(parent.d_sb, name)
	if &dentry == nil {
		return nil
	}
	dentry.d_flags = dentry.d_flags | DCACHE_RCUACCESS
	parent.d_lock.Lock()
	parent.__dget_dlock()
	
	dentry.d_parent = parent
	//list_add(&dentry.d_sisters, &parent.d_subdirs) //insert d_sisters into the parent's d_subdirs
	//fmt.Printf("parent d_subdirs before: %+v \n", stringifyList(&parent.d_subdirs))
	list_add(&list_node{data:dentry}, &parent.d_subdirs)
	//fmt.Printf("parent d_subdirs after: %+v \n", stringifyList(&parent.d_subdirs))

	//fmt.Printf("My name is %v, and my parent's name is %v \n", name, parent.d_name)

	parent.d_lock.Unlock()

	return dentry
}

func (d *dcache_l) __dget_dlock(dentry *Dentry) { //don't understand why this function is useful.
	dentry.d_lock.count++
}

func (d *dcache_l) d_set_d_op(dentry *Dentry, op *dentry_operations) {
	dentry.d_op = *op

	if op != nil {
		return
	}

	if &op.d_hash != nil {
		dentry.d_flags = dentry.d_flags | DCACHE_OP_HASH
	}
	if &op.d_compare != nil {
		dentry.d_flags = dentry.d_flags | DCACHE_OP_COMPARE
	}
	if &op.d_revalidate != nil {
		dentry.d_flags = dentry.d_flags | DCACHE_OP_REVALIDATE
	}
	if &op.d_weak_revalidate != nil {
		dentry.d_flags = dentry.d_flags | DCACHE_OP_WEAK_REVALIDATE
	}
	if &op.d_delete != nil {
		dentry.d_flags = dentry.d_flags | DCACHE_OP_DELETE
	}
	if &op.d_prune != nil {
		dentry.d_flags = dentry.d_flags | DCACHE_OP_PRUNE
	}
	if &op.d_real != nil {
		dentry.d_flags = dentry.d_flags | DCACHE_OP_REAL
	}
}

func (d *dcache_l) d_same_name(dentry *Dentry, parent *Dentry, name string) bool {
	//not allowing the DCACHE_OP_COMPARE flag.
	return (d.dentry_cmp(dentry, name) == true)
}

/*
func d_hash_and_lookup(dir *Dentry, qstr *stringPlusHash) *Dentry {
	toHash := getDentryStringUID(dir) + qstr.name
	qstr.hash = hashMeString(toHash)
	//Don't allow custom hashing.
	return d_lookup(dir, qstr)
}
*/

func (d *dcache_l) __d_lookup(parent *Dentry, name string) *Dentry {
	var toFindHash uint32 = getHashOfParentAndName(parent, name)
	//fmt.Printf("Retreival Hash: %+v", toFindHash)
	//fmt.Printf("dentry_hashtable: %+v \n", dentry_hashtable)
	var head *hlist_bl_head = d.getHomeListOfDentry(toFindHash) //the head with the correct hash.
	var node *hlist_bl_node
	var found *Dentry
	var dentry *Dentry

	node = head.first
	//fmt.Printf("Node: %+v", node)
	dentry = node.data.(*Dentry)

	for {
		if getHashOfDentry(dentry) != toFindHash {
			//fmt.Printf("Dentry hash Unequal \n")
			goto next
		}

		dentry.d_lock.Lock()

		if dentry.d_parent != parent {
			//fmt.Printf("Parent Unequal \n")
			goto next
		}

		if d.d_unhashed(dentry) {
			//fmt.Printf("Rip3 \n")
			goto next
		}

		if !d.d_same_name(dentry, parent, name) {
			//fmt.Printf("Name unequal \n")
			goto next
		}

		dentry.d_lock.count++
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

func (d *dcache_l) d_lookup(parent *Dentry, name string) *Dentry {
	var seq uint32
	var dentry *Dentry
	for {
		seq = d.rename_lock.read_seqbegin()
		dentry = d.__d_lookup(parent, name)
		if dentry != nil || !d.rename_lock.read_seqretry(seq) {
			break;
		}
	}

	return dentry
}

func (d *dcache_l) is_root(dentry *Dentry) bool{
	return dentry == dentry.d_parent
}

func (d *dcache_l) __d_drop(dentry *Dentry) { //requires d.d_lock
	if ! d.d_unhashed(dentry) {
		var b *hlist_bl_head

		if(d.is_root(dentry)) {
			b = &dentry.d_sb.s_anon //might get rid of this
		} else {
			b = d.getHomeListOfDentry(getHashOfDentry(dentry))
		}

		b.hlist_bl_lock()
		__hlist_bl_del(&dentry.d_masterListNode)
		dentry.d_masterListNode.pprev = nil
		b.hlist_bl_unlock()

		dentry.d_seq.write_seqlock_invalidate()
	}
}

func (d *dcache_l) d_drop(dentry *Dentry) {
	dentry.d_lock.Lock()
	d.__d_drop(dentry)
	dentry.d_lock.Unlock()
}

func (d *dcache_l) d_unhashed(dentry *Dentry) bool {
	return dentry.d_masterListNode.hlist_bl_unhashed()
}

func (d *dcache_l) d_is_negative(dentry *Dentry) bool {
	return dentry.d_type == DENTRY_MISS_TYPE
}

func (c *dcache_l) getSuperBlock() (*super_block){
	return &c.global_super_block
}

//dcache.h


const (
/* d_flags entries */
DCACHE_OP_HASH                 = 0x00000001
DCACHE_OP_COMPARE              = 0x00000002
DCACHE_OP_REVALIDATE           = 0x00000004
DCACHE_OP_DELETE               = 0x00000008
DCACHE_OP_PRUNE                = 0x00000010

DCACHE_DISCONNECTED            = 0x00000020
DCACHE_REFERENCED              = 0x00000040 /* Recently used, don't discard. */

DCACHE_CANT_MOUNT              = 0x00000100
DCACHE_GENOCIDE                = 0x00000200
DCACHE_SHRINK_LIST             = 0x00000400
DCACHE_OP_WEAK_REVALIDATE      = 0x00000800

DCACHE_NFSFS_RENAMED           = 0x00001000
DCACHE_COOKIE                  = 0x00002000 /* For use by dcookie subsystem */
DCACHE_FSNOTIFY_PARENT_WATCHED = 0x00004000
DCACHE_DENTRY_KILLED           = 0x00008000
DCACHE_MOUNTED                 = 0x00010000 /* is a mountpoint */
DCACHE_NEED_AUTOMOUNT          = 0x00020000 /* handle automount on this dir */
DCACHE_MANAGE_TRANSIT          = 0x00040000 /* manage transit from this dirent */
//DCACHE_MANAGED_DENTRY (DCACHE_MOUNTED|DCACHE_NEED_AUTOMOUNT|DCACHE_MANAGE_TRANSIT)

DCACHE_LRU_LIST                = 0x00080000

//DCACHE_ENTRY_TYPE              = 0x00700000
//DCACHE_MISS_TYPE               = 0x00000000 /* Negative dentry (maybe fallthru to nowhere) */
//DCACHE_WHITEOUT_TYPE           = 0x00100000 /* Whiteout dentry (stop pathwalk) */
//DCACHE_DIRECTORY_TYPE          = 0x00200000 /* Normal directory */
//DCACHE_AUTODIR_TYPE            = 0x00300000 /* Lookupless directory (presumed automount) */
//DCACHE_REGULAR_TYPE            = 0x00400000 /* Regular file type (or fallthru to such) */
//DCACHE_SPECIAL_TYPE            = 0x00500000 /* Other file type (or fallthru to such) */
//DCACHE_SYMLINK_TYPE            = 0x00600000 /* Symlink (or fallthru to such) */

DCACHE_MAY_FREE                = 0x00800000
DCACHE_FALLTHRU                = 0x01000000 /* Fall through to lower layer */
DCACHE_ENCRYPTED_WITH_KEY      = 0x02000000 /* dir is encrypted with a valid key */
DCACHE_OP_REAL                 = 0x04000000
DCACHE_PAR_LOOKUP              = 0x10000000 /* being looked up (with parent locked shared) */
DCACHE_DENTRY_CURSOR           = 0x20000000
)