package event

// Ascending is a slice of events that sorts in ascending chronological order
type Ascending []*T

func (ev Ascending) Len() no         { return len(ev) }
func (ev Ascending) Less(i, j no) bo { return *ev[i].CreatedAt < *ev[j].CreatedAt }
func (ev Ascending) Swap(i, j no)    { ev[i], ev[j] = ev[j], ev[i] }

// Descending sorts a slice of events in reverse chronological order (newest
// first)
type Descending []*T

func (e Descending) Len() no         { return len(e) }
func (e Descending) Less(i, j no) bo { return *e[i].CreatedAt > *e[j].CreatedAt }
func (e Descending) Swap(i, j no)    { e[i], e[j] = e[j], e[i] }
