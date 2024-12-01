package json

type I interface {
	Marshal(dst by) (b by)
	Unmarshal(dst by) (rem by, err er)
}

// An Object is an (not necessarily) ordered list of KeyValue.
type Object struct{ V []KeyValue }
