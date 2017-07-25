package swim

import (
	"bytes"
	"encoding/binary"
)

func converToBytes(obj interface{}, msgType uint8) []byte {
	buff := new(bytes.Buffer)
	buff.WriteByte(msgType)
	binary.Write(buff, binary.BigEndian, &obj)
	return buff.Bytes()
}

func convertToObj(message []byte) (interface{}, uint8) {
	var obj interface{}
	var msgType uint8
	buff := new(bytes.Buffer)
	msgBuf := new(bytes.Buffer)

	msgBuf.Write(message[:1])
	buff.Write(message[1:])

	binary.Read(buff, binary.BigEndian, &obj)
	binary.Read(msgBuf, binary.BigEndian, msgType)
	return &obj, msgType
}

func (s *Swim) isLocalNode(node *Node) (bool, int) {
	for i, localNode := range s.nodes {
		if localNode.Name == node.Name {
			return true, i
		}
	}
	return false, -1
}
