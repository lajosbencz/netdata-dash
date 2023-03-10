package agent

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
)

type Address struct {
	Host string `json:"host,omitempty"`
	Port int    `json:"port,omitempty"`
}

func (r Address) Format() string {
	return fmt.Sprintf("%s:%d", r.Host, r.Port)
}

type Config struct {
	HostName string  `json:"hostname,omitempty"`
	Realm    string  `json:"realm,omitempty"`
	Router   Address `json:"router,omitempty"`
	Netdata  Address `json:"netdata,omitempty"`
}

func DefaultConfig() *Config {
	return &Config{
		HostName: "localhost",
		Realm:    "netdata",
		Router: Address{
			Host: "localhost",
			Port: 16666,
		},
		Netdata: Address{
			Host: "localhost",
			Port: 19999,
		},
	}
}

func (r *Config) FromFile(jsonPath string) error {
	if _, err := os.Stat(jsonPath); err == nil {
		fh, err := os.Open(jsonPath)
		if err != nil {
			log.Fatalln(err)
		}
		if err := json.NewDecoder(fh).Decode(r); err != nil {
			log.Fatalln(err)
		}
	}
	return nil
}
