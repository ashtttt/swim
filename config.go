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
		ProbeTimeout:    3 * time.Second,
		Name:            hostname,
		BindAddr:        "0.0.0.0",
		BindPort:        8999,
	}
	return config

}
