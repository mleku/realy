package messages

type SubMessage struct {
	Id   int
	Name string
}

type Message struct {
	//M1 SubMessage
	//M2 *SubMessage
	SubMessage
}
