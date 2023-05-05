package conf

type RedisConfig struct {
	Network            string `yaml:"network"`
	Address            string `yaml:"address"`
	Password           string `yaml:"password"`
	MaxIdle            int    `yaml:"maxIdle"`
	IdleTimeoutSeconds int    `yaml:"idleTimeout"`
	// 连接池最大存活的连接数
	MaxActive int `yaml:"maxActive"`
	// 当连接数达到上限时，新的请求是等待还是立即报错.
	Wait bool `yaml:"wait"`
}

var defaultRedisConfig *RedisConfig

func GetDefaultRedisConfig() *RedisConfig {
	return defaultRedisConfig
}
