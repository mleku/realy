package number

import "fmt"

type List []no

// HasNumber returns true if the list contains a given number
func (l List) HasNumber(n no) (idx no, has bo) {
	for idx = range l {
		if l[idx] == n {
			has = true
			return
		}
	}
	return
}

func (l List) String() (s st) {
	s += "["
	for i := range l {
		if i > 0 {
			s += ","
		}
		s += fmt.Sprint(l[i])
	}
	s += "]"
	return
}
