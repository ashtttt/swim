package swim

import (
	"os"
	"time"
)

type Config struct {
	ProbingInterval time.Duration
	SubGroupSize    int
	ProbeTimeout    time.Duration

	Name     string
	BindAddr string
	BindPort int
}

func (c *Config) DefaultConfig() *Config {
	hostname, _ := os.Hostname()
	config := &Config{
		ProbingInterval: 1 * time.Second,
		ProbeTimeout:    500 * time.Millisecond,
		Name:            hostname,
		BindAddr:        "10.184.85.15",
		BindPort:        1023,
	}
	return config

}
