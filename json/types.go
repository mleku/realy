package json

type I interface {
	Marshal(dst B) (b B)
	Unmarshal(dst B) (rem B, err E)
}

// An Object is an (not necessarily) ordered list of KeyValue.
type Object struct{ V []KeyValue }
