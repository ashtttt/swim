package swim

import (
	"errors"
	"log"
	"net"
	"time"
)

type TransportConfig struct {
	BindAddr string
	BindPort int
}
type Transport struct {
	config *TransportConfig
	tcpCh  chan net.Conn
}

func NewTransport(config *TransportConfig) (*Transport, error) {

	if len(config.BindAddr) <= 0 {
		return nil, errors.New("bind address is required")
	}

	transport := &Transport{
		config: config,
		tcpCh:  make(chan net.Conn),
	}

	ip := net.ParseIP(config.BindAddr)
	listnerAddr := &net.TCPAddr{IP: ip, Port: config.BindPort}

	listner, err := net.ListenTCP("tcp", listnerAddr)
	log.Print(listnerAddr)
	if err != nil {
		log.Print(err)
		return nil, err
	}
	go transport.listen(listner)
	return transport, nil

}

func (t *Transport) listen(listner *net.TCPListener) {
	log.Print("listening")
	for {
		con, err := listner.AcceptTCP()
		if err != nil {
			log.Print("error accepting TCP connection: %v", err)
		}
		t.tcpCh <- con

	}
}

func (s *Transport) getDailer(node net.Addr, timeout time.Duration) (net.Conn, error) {
	dailer := net.Dialer{Timeout: timeout}
	conn, err := dailer.Dial("tcp", node.String())

	if err != nil {
		// TODO: Log Error
		return nil, err
	}
	return conn, nil
}
