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
	defaultSchedulerAppConfig = gConf.Scheduler
	defaultTriggerAppConfig = gConf.Trigger
	defaultMigratorAppConfig = gConf.Migrator
	defaultMySQLConfig = gConf.Mysql
	defaultRedisConfig = gConf.Redis
	defaultWebServerAppConf = gConf.WebServer
}

// gConf 兜底配置，即默认配置。后续配置文件会写入覆盖
var gConf GlobalConf = GlobalConf{
	Scheduler: &SchedulerAppConfig{
		// 分桶数量
		BucketsNum: 20,
		// 调度器获取分布式锁时初设的过期时间，单位：s
		TryLockSeconds: 70,
		// 调度器每次尝试获取分布式锁的时间间隔，单位：毫秒
		TryLockGapMilliSeconds: 100,
		// 时间片执行成功后，更新的分布式锁时间，单位：s
		SuccessExpireSeconds: 130,
	},

	Trigger: &TriggerAppConfig{
		// 触发器轮询定时任务 zset 的时间间隔，单位：s
		ZRangeGapSeconds: 1,
		// 并发协程数
		WorkersNum: 10000,
	},

	Migrator: &MigratorAppConfig{
		WorkersNum: 1000,
		// 一级每次迁移数据的时间间隔，单位：min
		MigrateStepMinutes: 60,
		// 迁移成功更新的锁过期时间，单位：min
		MigrateSuccessExpireMinutes: 120,
		// 迁移器获取锁时，初设的过期时间，单位：min
		MigrateTryLockMinutes: 20,
		// 迁移器提前将定时器数据缓存到内存中的保存时间，单位：min
		// 2 级迁移时间
		TimerDetailCacheMinutes: 2,
	},

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
}

type GlobalConf struct {
	Scheduler *SchedulerAppConfig `yaml:"scheduler"`
	Migrator  *MigratorAppConfig  `yaml:"*migrator"`
	Mysql     *MySQLConfig        `yaml:"mysql"`
	Redis     *RedisConfig        `yaml:"redis"`
	WebServer *WebServerAppConfig `yaml:"webservice"`
	Trigger   *TriggerAppConfig   `yaml:"trigger"`
}
