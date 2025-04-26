package primitive_alias

import "time"

type Aint int
type SliceInt []int
type Aints []Aint
type Astring string
type Astrings []Astring

type Tt struct {
	Id   int
	Name string
}

type Message struct {
	T2 Tt
	C1 SliceInt
	A1 Aint
	D1 Aints
	S1 Astrings
	T1 time.Time
}
