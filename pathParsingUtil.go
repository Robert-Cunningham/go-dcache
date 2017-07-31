package dcache2

import "strings"

func detectLastType(path string) (LastElementType) {
	var lastType LastElementType = LAST_NORMAL

	if path[0] == '.' {
		if len(path) > 1 && path[1] == '.' {
			lastType = LAST_DOUBLEDOT
		} else {
			lastType = LAST_DOT
		}
	}

	return lastType
}

func remove_leading_slashes(path string) string {
	if path == "" {
		return ""
	}
	location := 0
	for (location < len(path) && path[location] == '/' ) {location++}
	return path[location:]
}


func splitPath(path string) (string, string) {
	//fmt.Printf("splitting %v into %v and %v. \n", path, dirname(path), basename(path))
	return dirname(path), basename(path)
}

func basename(path string) string {
	return path[strings.LastIndex(path, "/") + 1:]
}

func dirname(path string) string {
	return path[:strings.LastIndex(path, "/") + 1] // x/y
}

func newPLS(parent *Dentry, root *Dentry, path string) *NewPathLookupState {
	cpls := NewPathLookupState{}
	cpls.currentDentry = parent
	cpls.root = root
	cpls.pathLeftToResolve = path
	return &cpls
}

type PathParser interface {
	resolvePath(path string) (*Dentry, PotentialError)
	getDcache() (dcache)
	setDcache(dcache)
}