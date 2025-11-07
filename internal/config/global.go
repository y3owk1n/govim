package config

var globalConfig *Config

// SetGlobal sets the global configuration that can be accessed from anywhere
func SetGlobal(cfg *Config) {
	globalConfig = cfg
}

// Global returns the global configuration
func Global() *Config {
	return globalConfig
}
