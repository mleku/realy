package slices

type SubMessage struct {
	Id   int
	Name string
}

//type SubMessage2 []SubMessage


type Message struct {
	M1 []SubMessage
//	M2 []*SubMessage
	//M3 SubMessages
	//M4 *SubMessages
//	SubMessage2
//	SubMessages2
}
