package dcache2

import (
	"fmt"
	"testing"
)

func TestNameSplitting(t *testing.T) {
	d, b := splitPath("test/path")
	if d != "test/" || b != "path" {
		t.Fail()
	}

	d, b = splitPath("/path")
	if d != "/" || b != "path" {
		t.Fail()
	}

	d, b = splitPath("/the/test/path")
	if d != "/the/test/" || b != "path" {
		t.Fail()
	}
}

func TestBasicFS(t *testing.T) {

	c := InitializeDcache()

	cd := createDirectory("/dankDir")
	printFSTree(c)
	if cd != SUCCESS {
		t.Fail()
	}

	cf := createFile("/dankDir/dankFile", "dankFileData")
	printFSTree(c)
	if cf != SUCCESS {
		t.Fail()
	}

	a, of := openFile("/dankDir/dankFile")
	if of != SUCCESS {
		t.Fail()
	}
	fmt.Printf("Opened File with data %v \n", (*a).(string))

	df := deleteFile("/dankDir/dankFile")
	printFSTree(c)
	if df != SUCCESS {
		t.Fail()
	}
}

func TestFolderDeletionFS(t *testing.T) {

	c := InitializeDcache()

	cd := createDirectory("/dankDir")
	printFSTree(c)
	if cd != SUCCESS {
		t.Fail()
	}

	cf := createFile("/dankDir/dankFile", "dankFileData")
	printFSTree(c)
	if cf != SUCCESS {
		t.Fail()
	}

	a, of := openFile("/dankDir/dankFile")
	if of != SUCCESS {
		t.Fail()
	}
	fmt.Printf("Opened File with data %v \n", (*a).(string))

	df := deleteFile("/dankDir")
	printFSTree(c)
	if df != SUCCESS {
		t.Fail()
	}
}
