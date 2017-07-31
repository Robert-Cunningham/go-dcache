package dcache2

import (
	"math/rand"
	"fmt"
)

func printDentryTree(c dcache, parent *Dentry, depth int) {
	var currentNode *list_node = &parent.d_subdirs
	var currentChild *Dentry;

	//fmt.Printf("current Node: %+v \n", currentNode)
	//fmt.Printf("current Child: %+v \n", currentChild)



	for {
		if(currentNode.next == nil) {
			return
		} else if(currentNode.data == nil) {
			goto next
		} else {
			 currentChild = currentNode.data.(*Dentry)
		}
		for i := 0; i < depth; i++ {
			fmt.Printf("  ")
		}
		fmt.Printf("|-")
		fmt.Printf("%v ", currentChild.d_name)
		if(c.d_is_negative(currentChild)) {
			fmt.Printf(" [NEGATIVE] ")
		}
		fmt.Println()
		printDentryTree(c, currentChild, depth + 1)

		next:
		currentNode = currentNode.next
		//currentChild = currentNode.data.(*Dentry)
	}

}

func printFSTree(c dcache) {
	fmt.Printf("CURRENT STATE OF FILESYSTEM\n---------------------------\n")
	printDentryTree(c, c.getSuperBlock().grandparent, 0)
}

func generateRandomFSTree(dentryCount int) (dcache_rw) {
	//rand.Seed(time.Now().UnixNano())

	c := dcache_rw_init()

	iterations := dentryCount

	var dentries []*Dentry = make([]*Dentry, dentryCount)
	dentries[0] = c.getSuperBlock().grandparent

	for i := 1; i < iterations; i++ {
		parent := dentries[rand.Intn(i)]
		name := getRandomString(3)
		myDentry := c.d_alloc(parent, name)

		myInode := newInodeOld(fmt.Sprintf("Very Important Data:%v", i))
		c.d_add(myDentry, &myInode)
		dentries[i] = myDentry

		//fmt.Printf("Added a node with the name %v, whose parent's subdirs are %v \n", name, stringifyList(&parent.d_subdirs))
	}

	fmt.Println("Overwrote FS with a random FS tree.")

	return *c
}

func getRandomDentryFromPath(c dcache) *Dentry {
	_, d, _ := generateRandomPath(c, false)
	return d
}

func generateRandomPath(c dcache, canFail bool) (string, *Dentry, PotentialError) {
	currentError := SUCCESS
	currentDentry := c.getSuperBlock().grandparent
	path := ""
	depth := 0
	for {
		r := rand.Float32()
		if r < 0.65 {
			subdirs := convertInterfaceArrayToDentryArray(removeNilFromSlice(sliceifyList(&currentDentry.d_subdirs)))
			if len(subdirs) > 0 {
				newDirectory := subdirs[rand.Intn(len(subdirs))]
				path = path + "/" + newDirectory.d_name
				depth++
				currentDentry = newDirectory
			}
		} else if (r < 0.80 && depth > 0) {
			currentDentry = currentDentry.d_parent
			path = path + "/.."
			depth--
		} else if (r < 0.85) {
			path = path + "/."
		} else if (r < 0.90){
			path = path + "//"
		} else if (r < 0.93 && canFail) {
			path = path + "/" + getRandomString(3)
			if(currentError == SUCCESS) {
				currentError = ERROR_FAILED_TO_FIND
			}
		} else {
			break
		}
	}

	//returns a path, the dentry it's supposed to lead to, and the error state that ought to arise from it.
	return path, currentDentry, currentError }

//Thanks http://stackoverflow.com/questions/22892120/how-to-generate-a-random-string-of-a-fixed-length-in-golang (laziness)

var letterBytes = "abcdefghijklmnopqrstuvwxyz"

func getRandomString(length uint) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = letterBytes[rand.Int63() % int64(len(letterBytes))]
	}
	return string(b)
}