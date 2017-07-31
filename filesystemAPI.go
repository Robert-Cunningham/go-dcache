package dcache2

import "fmt"

var pp ThePathParser = ThePathParser{}

func createFile(path string, data interface{}) (PotentialError) {
	dir, base := splitPath(path)
	parent, e := pp.resolvePath(dir)

	if(e != SUCCESS) {
		return e
	}

	myDentry := pp.getDcache().d_alloc(parent, base)
	pp.getDcache().d_add(myDentry, newInode(data))

	return SUCCESS
}

func createDirectory(path string) (PotentialError) {
	dir, base := splitPath(path)
	parent, e := pp.resolvePath(dir)

	if(e != SUCCESS) {
		return e
	}

	myDentry := pp.getDcache().d_alloc(parent, base)
	myDentry.setAsDirectory()
	myInode := newInode("I am an inode directory")
	pp.getDcache().d_add(myDentry, myInode)

	return SUCCESS
}

func deleteFile(path string) (PotentialError) {
	dentry, e := pp.resolvePath(path)

	if(e != SUCCESS) {
		return e
	}

	pp.getDcache().d_delete(dentry)
	fmt.Printf("DENTRY: %+v \n", dentry)

	return SUCCESS
}

func openFile(path string) (*interface{}, PotentialError) {
	dentry, de := pp.resolvePath(path)
	data, ie := dentry.getInodeData()

	if(de != SUCCESS) {
		return nil, de
	} else if ie != SUCCESS {
		return nil, ie
	} else {
		return data, SUCCESS
	}


}

func overwriteFile(path string, data interface{}) (PotentialError) {
	dentry, de := pp.resolvePath(path)

	if(de != SUCCESS) {
		return de
	}

	ie := dentry.setInodeData(&data)

	if(ie != SUCCESS) {
		return ie
	} else {
		return SUCCESS
	}
}