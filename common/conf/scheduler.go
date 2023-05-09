package conf

type SchedulerAppConfig struct {
	BucketsNum             int `yaml:"bucketsNum"`
	TryLockSeconds         int `yaml:"tryLockSeconds"`
	TryLockGapMilliSeconds int `yaml:"tryLockGapMilliSeconds"`
	SuccessExpireSeconds   int `yaml:"successExpireSeconds"`
}

var defaultSchedulerAppConfig *SchedulerAppConfig

func GetDefaultSchedulerAppConfig() *SchedulerAppConfig {
	return defaultSchedulerAppConfig
}
