package conf

import (
	"fmt"
	"github.com/spf13/viper"
)

func init() {
	viper.SetConfigName("config") // name of config file (without extension)
	viper.SetConfigType("yaml")   // REQUIRED if the config file does not have the extension in the name
	viper.AddConfigPath(".")      // path to look for the config file in

	// 读取配置文件
	err := viper.ReadInConfig() // Find and read the config file
	if err != nil {             // Handle errors reading the config file
		panic(fmt.Errorf("fatal error config file: %w", err))
	}

	// 序列化到当前包变量中
	err = viper.Unmarshal(&gConf)
	if err != nil {
		panic(fmt.Errorf("fatal error config file: %w", err))
	}

	// 序列化到各包变量中，方便 provide 到 contain 中
	defaultMySQLConfig = gConf.Mysql
	defaultRedisConfig = gConf.Redis
	defaultPoolConfig = gConf.Pool
}

// gConf 兜底配置，即默认配置。后续配置文件会写入覆盖
var gConf GlobalConf = GlobalConf{
	Redis: &RedisConfig{
		Network: "tcp",
		// 最大空闲连接数
		MaxIdle: 2000,
		// 空闲连接超时时间，单位：s
		IdleTimeoutSeconds: 30,
		// 连接池最大存活的连接数
		MaxActive: 1000,
		// 当连接数达到上限时，新的请求是等待还是立即报错
		Wait: true,
	},
	Mysql: &MySQLConfig{
		MaxOpenConns: 100,
		MaxIdleConns: 50,
	},
	Pool: &PoolConfig{
		Size:          10000,
		ExpireSeconds: 60,
		nonBlocking:   false,
	},
}

type GlobalConf struct {
	Mysql *MySQLConfig `yaml:"mysql"`
	Redis *RedisConfig `yaml:"redis"`
	Pool  *PoolConfig  `yaml:"pool"`
}
