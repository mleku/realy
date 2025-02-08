package number

import "fmt"

type List []int

func (l List) Len() int           { return len(l) }
func (l List) Less(i, j int) bool { return l[i] < l[j] }
func (l List) Swap(i, j int)      { l[i], l[j] = l[j], l[i] }

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

func (l List) String() (s string) {
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
