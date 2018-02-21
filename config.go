package lora

const (
	defaultPort       = 48071
	defaultDistroHost = "127.0.0.1"
)

type Config struct {
	Port       int
	DistroHost string
}

var cfg Config

func GetDefaultConfig() Config {
	return Config{
		Port:       defaultPort,
		DistroHost: defaultDistroHost,
	}
}
