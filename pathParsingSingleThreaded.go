package dcache2

import (
	"fmt"
	"strings"
)

/*
	link_path_walk (resolvePath)
	--walk_component (walk_component)
	----lookup_fast (resolveNextDentryFast)
	------d_lookup_rcu
	----lookup_slow (resolveNextDentrySlow)
	------d_lookup
*/

type ThePathParser struct {
	dc dcache
}

func (s *ThePathParser) getDcache() dcache {
	return s.dc
}

func (s *ThePathParser) setDcache(newdcache dcache) {
	s.dc = newdcache
}

func (s *ThePathParser) resolvePathFromLocation(parent *Dentry, root *Dentry, path string) (*Dentry, PotentialError) {
	return s.resolvePathPLS(newPLS(parent, root, path))
}

func (s ThePathParser) resolvePath(path string) (*Dentry, PotentialError) {
	return s.resolvePathFromLocation(s.getDcache().getSuperBlock().grandparent, s.getDcache().getSuperBlock().grandparent, path)
}

func (s *ThePathParser) resolvePathPLS(cpls *NewPathLookupState) (*Dentry, PotentialError) {
	for {
		cpls.pathLeftToResolve = remove_leading_slashes(cpls.pathLeftToResolve)
		if cpls.pathLeftToResolve == "" {
			break
		}

		e := s.walk_component(cpls) //resolve the next bit of the path.

		if e != SUCCESS {
			return nil, e
		}
	}

	return cpls.currentDentry, SUCCESS
}

func (s *ThePathParser) walk_component(cpls *NewPathLookupState) PotentialError {
	var e PotentialError

	//fmt.Println("Attmpeting to resolve", cpls.pathLeftToResolve)

	lastType := detectLastType(cpls.pathLeftToResolve)

	if lastType == LAST_DOT || lastType == LAST_DOUBLEDOT {
		e = s.handle_dots(cpls, lastType)
	} else {
		e = s.resolveNextDentryFast(cpls)

		if e != SUCCESS {
			_, e = s.resolveNextDentrySlow(cpls)
		}
	}

	if strings.Contains(cpls.pathLeftToResolve, "/") {
		cpls.pathLeftToResolve = cpls.pathLeftToResolve[strings.Index(cpls.pathLeftToResolve, "/"):]
	} else {
		cpls.pathLeftToResolve = ""
	}
	cpls.depth++

	return e
}

func (s *ThePathParser) resolveNextDentryFast(cpls *NewPathLookupState) PotentialError { //lookup_fast, TODO
	return ERROR_ILLEGAL
}

func (s *ThePathParser) resolveNextDentrySlow(cpls *NewPathLookupState) (*Dentry, PotentialError) {
	toFind := strings.Split(cpls.pathLeftToResolve, "/")[0]
	//fmt.Printf("Searching for %v in %v \n", toFind, cpls.currentDentry)
	found := pp.getDcache().d_lookup(cpls.currentDentry, toFind)
	//fmt.Printf("Hit: %v \n", found)

	if found == nil {
		return nil, ERROR_FAILED_TO_FIND
	} else {
		cpls.currentDentry = found
		return found, SUCCESS
	}

}

func switchToREFWalk(cpls *NewPathLookupState, dentry *Dentry, seq uint32) int { //unlazy_walk equivalent
	if !cpls.usingRCU {
		fmt.Println("DISASTER! Switched to REF without being in RCU mode!!")
	}

	cpls.usingRCU = false
	if cpls.currentDentry.isDead {
		return -ERROR_LINUX_ECHILD
	}

	if dentry == nil {
		if cpls.currentDentry.d_seq.read_seqretry(cpls.seq) {
			return -ERROR_LINUX_ECHILD
		}
	} else {
		if dentry.isDead {
			return -ERROR_LINUX_ECHILD
		}

		if dentry.d_seq.read_seqretry(seq) {
			//delete dentry
			//dentry.dcache_parent.dput(dentry)
			return -ERROR_LINUX_ECHILD
		}
	}

	if !legitimize_path(cpls, dentry, seq) {
		//dentry.dcache_parent.dput(dentry)
		return -ERROR_LINUX_ECHILD
	}

	return 0
}

func legitimize_path(cpls *NewPathLookupState, dentry *Dentry, seq uint32) bool {
	if dentry.isDead {
		dentry = nil
		return false
	}

	return !dentry.d_seq.read_seqretry(seq)
}

func (s *ThePathParser) handle_dots(cpls *NewPathLookupState, t LastElementType) PotentialError {

	if t == LAST_DOUBLEDOT {
		if cpls.root == cpls.currentDentry {
			return ERROR_ILLEGAL
		}

		cpls.currentDentry = cpls.currentDentry.d_parent
		return SUCCESS

	} else {
		return SUCCESS
	}
}
