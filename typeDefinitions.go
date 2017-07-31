package dcache2

import "fmt"

type InodeType uint8

const (
	INODE_DIRECTORY_TYPE InodeType = iota + 1
	INODE_REGULAR_TYPE
)

type PathLookupState struct { //equivalent of nameidata
	currentDentry Dentry
			      //path string
	root Dentry
	flags int32
	depth int
	last_type LastElementType
	last_component_name_and_hash stringPlusHash //equivalent of last
	seq uint
}

/*
func newStringPlusHash(s string) (stringPlusHash) {
	return stringPlusHash{hash_string(s, ), uint32(len(s)), s}
}
*/

type stringPlusHash struct { //qstr equivalent. length is unnecessary since Go keeps track of it for us.
	hash uint32
	string_len uint32 //equivalent to hash_len
	name string
}


type d_revalidate_func func(*Dentry, uint)
type d_weak_revalidate_func func(*Dentry, uint)
type d_hash_func func(*Dentry, stringPlusHash)
type d_compare_func func(*Dentry, *Dentry, uint, string, stringPlusHash)
type d_delete_func func(*Dentry)
type d_init_func func(*Dentry) int
type d_release_func func(*Dentry)
type d_prune_func func(*Dentry)
type d_iput_func func(*Dentry)
type d_dname_func func(*Dentry) (string)
type d_manage_func func(*Dentry, bool) (int)
type d_real_func func(*Dentry, *Inode, uint)

type dentry_operations struct {
	d_revalidate d_revalidate_func
	d_weak_revalidate d_weak_revalidate_func
	d_hash d_hash_func
	d_compare d_compare_func
	d_delete d_delete_func
	d_init d_init_func
	d_release d_release_func
	d_prune d_prune_func
	d_iput d_iput_func
	d_dname d_dname_func
	d_manage d_manage_func
	d_real d_real_func
}

type dentryType uint8

const (
	DENTRY_REGULAR_TYPE dentryType = iota + 1
	DENTRY_DIRECTORY_TYPE
	DENTRY_MISS_TYPE
)

type Dentry struct {

	d_flags          uint
	d_seq            SeqLock
	d_masterListNode hlist_bl_node //lookup hash list d_hash equiv.
	d_parent         *Dentry
	d_name           string
	d_inode          *Inode

	d_iname          string
	d_lock           LockRef
	d_op             dentry_operations
	d_sb             *super_block
	d_time           uint

	d_lru            list_node
				       //d_wait *wait_queue_head_t

	//d_sisters        list_node     //equivalent of d_child
	d_subdirs        list_node

				       //d_alias hlist_node //who else points to the same inode?
	d_in_lookup_hash hlist_bl_node
				       //d_rcu rcu_head

	d_type           dentryType

	dcache_parent dcache

	isDead		bool
}


func (d Dentry) String() string {
	return fmt.Sprintf("[Dentry; Name: %v, Parent: %v]", d.d_name, d.d_parent.d_name)
}


type super_block struct {
	grandparent *Dentry
	s_anon hlist_bl_head
	s_d_op dentry_operations //Where does this come from?
}

func newInodeOld(name string) Inode {
	return Inode{SpinLock{}, 0, false, name, nil, SpinLock{}}
}

func newInode(data interface{}) *Inode {
	return &Inode{SpinLock{}, 0, false, "", data, SpinLock{}}
}

type Inode struct {
	i_lock        SpinLock
	i_mode        InodeType
	isDeleted     bool
	name_outdated string
	data interface{}

	io_lock       SpinLock
}

type LastElementType uint8

const (
	LAST_DOT LastElementType = iota + 1
	LAST_DOUBLEDOT
	LAST_NORMAL
)

type NewPathLookupState struct { //equivalent of nameidata
	currentDentry *Dentry
	root *Dentry
	//flags int32
	usingRCU bool
	depth int
	last_type LastElementType
	pathLeftToResolve string
	seq uint32
}

type PotentialError uint8

const (
	SUCCESS PotentialError = iota + 1
	ERROR_FAILED_TO_FIND
	ERROR_ILLEGAL
	ERROR_COMBO
)

const (
	ERROR_LINUX_ECHILD = 1
)

func (a *PotentialError) or(b *PotentialError) (PotentialError){
	if(*a == SUCCESS && *b == SUCCESS) {
		return SUCCESS
	} else if (*a != SUCCESS) {
		return *a
	} else if (*b != SUCCESS) {
		return *b
	} else {
		fmt.Printf("WTF")
		return ERROR_ILLEGAL
	}
}

/*
func (a *PotentialError) foldIn (b *PotentialError) {
	b = b.or(a)
}
*/