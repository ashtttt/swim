package swim

type ping struct {
	SeqNo int
}

type ackMessage struct {
	SeqNo  int
	PayLod []*Node
}
