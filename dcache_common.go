package dcache2

import (
	"strconv"
)

const dhash_entries = 1000    //linux compilation option
const L1_CACHE_BYTES = 131072 //128k on FX-8350
const IN_LOOKUP_SHIFT = 10
const DNAME_INLINE_LEN = 50 //don't make the name any longer than this.

const DENTRY_HASHTABLE_SIZE = 1000

//var dc dcache = dcache_rw_init()

type dcache interface {
	//dcache_init()
	d_lookup(parent *Dentry, name string) *Dentry
	d_add(dentry *Dentry, inode *Inode)
	d_is_negative(*Dentry) bool
	getSuperBlock() *SuperBlock
	d_delete(*Dentry)
	NewDentry(parent *Dentry, name string) *Dentry
}

func getDentryStringUID(dentry *Dentry) string {
	//return getStringAddressOfDentry(dentry)
	return strconv.Itoa(dentry.d_uuid)
}

func getDentryIntUID(dentry *Dentry) uint64 {
	//return getAddressOfDentry(dentry)
	return uint64(dentry.d_uuid)
}

func getHashOfDentry(dentry *Dentry) uint32 {
	return getHashOfParentAndName(dentry.d_parent, dentry.d_name)
}

func getHashOfParentAndName(parent *Dentry, name string) uint32 {
	return hash_string(name, getDentryStringUID(parent))
}

func (d *Dentry) setInodeData(data *interface{}) PotentialError {
	d.d_inode = newInode(data)
	return SUCCESS
}

type dentryTranformation func(*Dentry)

func applyToDentryTreeHelper(transformation dentryTranformation, parent *Dentry, depth int) {
	var currentNode *list_node = &parent.d_subdirs
	var currentChild *Dentry

	for {
		if currentNode.next == nil {
			return
		} else if currentNode.data == nil {
			goto next
		} else {
			currentChild = currentNode.data.(*Dentry)
		}

		applyToDentryTreeHelper(transformation, currentChild, depth+1)

		transformation(currentChild)

	next:
		currentNode = currentNode.next
		//currentChild = currentNode.data.(*Dentry)
	}

}

func applyToDentryTree(transformation dentryTranformation, dentry *Dentry) {
	applyToDentryTreeHelper(transformation, dentry, 0)
	transformation(dentry) //allow destructive tranformations
}

func (d *Dentry) setAsDirectory() {
	d.d_type = DENTRY_DIRECTORY_TYPE
	//d.d_inode.i_mode = INODE_DIRECTORY_TYPE
}

func (d *Dentry) getInodeData() (*interface{}, PotentialError) {
	if d.d_inode != nil && !d.d_inode.isDeleted {
		return &d.d_inode.data, SUCCESS
	} else {
		return nil, ERROR_FAILED_TO_FIND
	}

}

const (
	DCACHE_RCUACCESS  = 0x00000080 /* Entry has ever been RCU-visible */
	DCACHE_PAR_LOOKUP = 0x00000800
)
