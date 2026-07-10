package config

// ServerConfig holds runtime settings for the Redis-compatible server. #genai
type ServerConfig struct {
	Host     string
	Port     int
	MaxKeys  int
	AOF_FILE string
}

// Config is the active server configuration, set via ForceInit and flags.
var Config *ServerConfig

// ForceInit assigns the global Config pointer before flags are bound.
func ForceInit(cfg *ServerConfig) {
	Config = cfg
	if Config.MaxKeys <= 0 {
		Config.MaxKeys = 5
	}
	if Config.AOF_FILE == "" {
		Config.AOF_FILE = "./mastr.aof"
	}
}
