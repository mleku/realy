package unix

import (
	"time"

	"realy.mleku.dev/ints"
)

type Time struct{ time.Time }

func Now() *Time { return &Time{Time: time.Now()} }

func (u *Time) MarshalJSON() (b []byte, err error) {
	b = ints.New(u.Time.Unix()).Marshal(b)
	return
}

func (u *Time) UnmarshalJSON(b []byte) (err error) {
	t := ints.New(0)
	_, err = t.Unmarshal(b)
	u.Time = time.Unix(int64(t.N), 0)
	return
}
