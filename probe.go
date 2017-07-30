package swim

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"time"
)

const (
	pingMsgType uint8 = 1
	ackMsgType  uint8 = 2
)

func (s *Swim) initProbing() {
	ticker := time.NewTicker(s.config.ProbingInterval)
	s.ticker = ticker
	for _ = range s.ticker.C {
		node := s.nextTarget()
		if node != nil {
			fmt.Printf("Nex target %s \n", node.Addr)
			go s.probe(node)
		}
	}
}

func (s *Swim) probe(node *Node) error {
	if len(node.Addr) <= 0 {
		return errors.New("Node address cant be empty")
	}
	hb := ping{SeqNo: s.nextSeqNo(), Name: node.Name}

	buff := serialize(&hb, pingMsgType)

	addr, _ := net.ResolveTCPAddr("tcp", node.Address())
	conn, err := s.transport.getDailer(addr, s.config.ProbeTimeout)
	if err != nil {
		//TODO: may be node is down -  try to se
		return err
	}

	defer conn.Close()

	// send ping message to connection
	err = s.sendMessage(buff, conn)

	// read the response from connection
	msgType, msgBytes, err := s.readMessage(conn)

	if msgType == ackMsgType {
		ack := new(ackMessage)
		err = deserialize(msgBytes, ack)
		if err != nil {
			//TODO: Log errors
			return err
		}
		if ack.SeqNo == hb.SeqNo {
			ip, _, _ := net.SplitHostPort(conn.RemoteAddr().String())
			err := s.setAlive(
				&Node{
					Name:   hb.Name,
					Addr:   ip,
					Port:   s.config.BindPort,
					Status: "alive",
				})
			err = s.handleAck(*ack)
			if err != nil {
				//TODO :  Log errors
				return err
			}
		}
	}

	return nil
}

func (s *Swim) sendMessage(message []byte, conn net.Conn) error {

	// message lenght
	ln := make([]byte, 4)
	binary.BigEndian.PutUint32(ln, uint32(len(message)))

	//write the lenght
	_, err := conn.Write(ln)

	_, err = conn.Write(message)
	if err != nil {
		return err
	}
	return nil
}

func (s *Swim) readMessage(conn net.Conn) (uint8, []byte, error) {

	data := bytes.NewBuffer(nil)

	_, err := io.CopyN(data, conn, 4)
	if err != nil {
		return 0, nil, err
	}

	msgLen := binary.BigEndian.Uint32(data.Bytes()[:4])

	_, err = io.CopyN(data, conn, int64(msgLen))
	if err != nil {
		return 0, nil, err
	}

	message := data.Bytes()[4:]
	msgType := uint8(message[0])
	return msgType, message[1:], nil
}

func (s *Swim) handleAck(ack ackMessage) error {

	for _, remoteNode := range ack.PayLod {
		status, i := s.isLocalNode(remoteNode)
		if status {
			s.nodes = append(s.nodes[:i], s.nodes[i+1:]...)
			s.nodes = append(s.nodes, remoteNode)
		} else {
			s.nodes = append(s.nodes, remoteNode)
		}
	}
	log.Print(s.nodes[0])
	return nil
}

func (s *Swim) setAlive(node *Node) error {
	staus, index := s.isLocalNode(node)
	if staus {
		localNode := s.nodes[index]
		localNode.Status = node.Status
	} else {
		s.nodes = append(s.nodes, node)
	}
	return nil
}
