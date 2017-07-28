package swim

import (
	"bytes"
	"encoding/gob"
)

func serialize(obj interface{}, msgType uint8) []byte {
	buf := bytes.NewBuffer(nil)
	buf.WriteByte(uint8(msgType))
	encoder := gob.NewEncoder(buf)
	encoder.Encode(obj)
	return buf.Bytes()

}

func deserialize(message []byte, obj interface{}) error {
	buff := bytes.NewBuffer(message)
	decoder := gob.NewDecoder(buff)
	err := decoder.Decode(obj)
	if err != nil {
		return err
	}
	return nil
}

func (s *Swim) isLocalNode(node *Node) (bool, int) {
	for i, localNode := range s.nodes {
		if localNode.Name == node.Name {
			return true, i
		}
	}
	return false, -1
}
