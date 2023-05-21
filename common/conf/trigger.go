package conf

type TriggerAppConfig struct {
	ZRangeGapSeconds int `yaml:"zrangeGapSeconds"`
	WorkersNum       int `yaml:"workersNum"`
}

var defaultTriggerAppConfig *TriggerAppConfig

func GetDefaultTriggerAppConfig() *TriggerAppConfig {
	return defaultTriggerAppConfig
}
