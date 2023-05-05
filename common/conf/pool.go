package conf

type PoolConfig struct {
	Size          int  `yaml:"size"`
	ExpireSeconds int  `yaml:"expireSeconds"`
	nonBlocking   bool `yaml:"nonBlocking"`
}

var defaultPoolConfig *PoolConfig

func GetDefaultPoolConfig() *PoolConfig {
	return defaultPoolConfig
}
