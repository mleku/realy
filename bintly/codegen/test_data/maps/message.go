package maps

type SubMessage struct {
	Id   int
	Name string
}


type M1 map[string][]*SubMessage

type Message struct {
	//M1 map[string]SubMessage
	//M1 map[string]*SubMessage
	//M1 map[string][]SubMessage
	//M1 map[string][]*SubMessage
	 M1
}
