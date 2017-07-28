package swim

type ping struct {
	Name  string
	SeqNo int
}

type ackMessage struct {
	Name   string
	SeqNo  int
	PayLod []*Node
}
