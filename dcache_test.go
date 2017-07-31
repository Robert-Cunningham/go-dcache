package dcache2

import (
	"testing"
	"fmt"
)

func TestApplyToDentryTree(t *testing.T) {
	c := dcache_rw_init()

	generateRandomFSTree(100)
	//printFSTree()
	applyToDentryTree(func (d *Dentry) {d.d_iname = "AFFECTED"}, c.getSuperBlock().grandparent)
	//printFSTree()
	for i := 0; i < 100; i++ {
		d := getRandomDentryFromPath(c)
		if(d.d_iname != "AFFECTED") {
			fmt.Printf("FALIED: %v, %v \n", d, d.d_iname)
			t.Fail()
		}
	}
}

func TestDcacheAddingDentry(t *testing.T) {
	c := dcache_rw_init()

	myDentry := c.d_alloc(c.global_super_block.grandparent, "babyDentry")

	myInode := newInodeOld("Very Important Data")
	c.d_add(myDentry, &myInode)

	returned_from_dead := c.d_lookup(c.global_super_block.grandparent, "babyDentry")
	//fmt.Printf("Returned from the dead: %+v \n", returned_from_dead.d_inode.data)
	if(returned_from_dead.d_inode.name_outdated != "Very Important Data") {
		t.Fail()
	}
}

func TestDcacheAddingDentries(t *testing.T) {
	c := dcache_rw_init()

	iterations := 10000

	for i := 0; i < iterations; i++ {

		myDentry := c.d_alloc(c.global_super_block.grandparent, fmt.Sprintf("babyDentry:%v", i))

		myInode := newInodeOld(fmt.Sprintf("Very Important Data:%v", i))
		c.d_add(myDentry, &myInode)
	}

	for i := 0; i < iterations; i++ {
		returned_from_dead := c.d_lookup(c.global_super_block.grandparent, fmt.Sprintf("babyDentry:%v", i))
		if(returned_from_dead.d_inode.name_outdated != fmt.Sprintf("Very Important Data:%v", i)) {
			t.Fail()
		}
	}
}

func TestDcacheDeletingDentry(t *testing.T) {
	c := dcache_rw_init()

	myDentry := c.d_alloc(c.global_super_block.grandparent, "babyDentry")

	myInode := newInodeOld("Very Important Data")
	c.d_add(myDentry, &myInode)

	should_be_alive := c.d_lookup(c.global_super_block.grandparent, "babyDentry")

	c.d_delete(myDentry)

	should_be_dead := c.d_lookup(c.global_super_block.grandparent, "babyDentry")

	//fmt.Printf("Should be dead: %v", should_be_dead)

	if(should_be_alive == should_be_dead) {
		t.Fail()
	}
}

func TestPathResolution(t *testing.T) {
	c := generateRandomFSTree(100)
	//fmt.Printf("GSB:GP %+v \n", &global_super_block.grandparent)
	//fmt.Printf("GSB:GP:SUBDIRS %+v \n", global_super_block.grandparent.d_subdirs)
	//fmt.Printf("%+v \n", global_super_block.grandparent.d_subdirs.data.(*Dentry))
	printFSTree(&c)

	iterations := 10

	for i := 0; i < iterations; i++ {
		path, desiredResult, desiredError := generateRandomPath(&c, true)
		actualResult, actualError := pp.resolvePath(path)

		if desiredError == actualError && (actualError != SUCCESS || desiredResult.d_name == actualResult.d_name) {
			fmt.Printf("Correctly resolved %v. \n", path)
		} else {
			fmt.Printf("Did not resolve %v correctly. \n", path)
			t.Fail()
		}
	}
	//result := resolvePath(global_super_block.grandparent, global_super_block.grandparent, "jnn/twt/hxt/../ilf/bgp").d_name
	//fmt.Printf("Result: %v \n", result)
}