package dcache2

import (
	"strings"
)

func (p Path) simplifyPath(pwd Path) Path { //gets rid of any dot or doubledots.
	simplified := p.convertToAbsolute(pwd)
	s := strings.Split(simplified.original, "/")
	for i := 0; i < len(s); i++ {
		if s[i] == "." {
			s = append(s[:i], s[i+1:]...)
			i--
		}
	}
	for i := 0; i < len(s); i++ {
		if s[i] == ".." {
			s = append(s[:i-1], s[i+1:]...)
			i -= 2
		}
	}
	return Path{strings.Join(s, "/")}
}

func (p Path) convertToAbsolute(pwd Path) Path {
	if p.isAbsolute() {
		return p
	}

	return Path{pwd.original + "/" + p.original}

}

func (p Path) isRelative() bool {
	return !p.isAbsolute()
}

func (p Path) isAbsolute() bool {
	return strings.HasPrefix(p.original, "/")
}

type Path struct {
	original string
}
