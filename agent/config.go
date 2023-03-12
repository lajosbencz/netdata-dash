package agent

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/lajosbencz/netdata-dash/utils"
)

type Address struct {
	Host string `json:"host,omitempty"`
	Port int    `json:"port,omitempty"`
}

func (r Address) Format() string {
	return fmt.Sprintf("%s:%d", r.Host, r.Port)
}

type Config struct {
	HostName string              `json:"hostname,omitempty"`
	HostTags utils.StringsUnique `json:"host_tags,omitempty"`
	Realm    string              `json:"realm,omitempty"`
	Dash     Address             `json:"dash,omitempty"`
	Netdata  Address             `json:"netdata,omitempty"`
}

func DefaultConfig() *Config {
	return &Config{
		HostName: "localhost",
		HostTags: utils.StringsUnique{},
		Realm:    "netdata",
		Dash: Address{
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
			return err
		}
		if err := json.NewDecoder(fh).Decode(r); err != nil {
			return err
		}
	}
	return nil
}
