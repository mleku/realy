package number

import "fmt"

type List []int

// HasNumber returns true if the list contains a given number
func (l List) HasNumber(n int) (idx int, has bool) {
	for idx = range l {
		if l[idx] == n {
			has = true
			return
		}
	}
	return
}

func (l List) String() (s S) {
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
