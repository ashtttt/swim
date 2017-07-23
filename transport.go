package swim

import (
	"errors"
	"log"
	"net"
)

type TransportConfig struct {
	BindAddr string
	BindPort int
	logger   *log.Logger
}
type Transport struct {
	config *TransportConfig
	tcpCh  chan net.Conn
	logger *log.Logger
}

func NewTransport(config *TransportConfig) (*Transport, error) {

	if len(config.BindAddr) <= 0 {
		return nil, errors.New("bind address is required")
	}

	transport := &Transport{
		config: config,
		tcpCh:  make(chan net.Conn),
		logger: config.logger,
	}

	ip := net.ParseIP(config.BindAddr)
	listnerAddr := &net.TCPAddr{IP: ip, Port: config.BindPort}

	listner, err := net.ListenTCP("tcp", listnerAddr)
	if err != nil {
		log.Print(err)
		return nil, err
	}
	log.Print("aftr listerner")
	go transport.listen(listner)
	return transport, nil

}

func (t *Transport) listen(listner *net.TCPListener) {
	log.Print("listineting")
	for {
		con, err := listner.AcceptTCP()
		if err != nil {
			log.Print("error accepting TCP connection: %v", err)
		}
		t.tcpCh <- con

	}
}
