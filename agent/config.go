package agent

type Config struct {
	RouterHost  string
	RouterPort  int
	NetdataHost string
	NetdataPort int
}

func DefaultConfig() *Config {
	return &Config{
		RouterHost:  "localhost",
		RouterPort:  16666,
		NetdataHost: "localhost",
		NetdataPort: 19999,
	}
}
