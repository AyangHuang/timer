package conf

type WebServerAppConfig struct {
	Port int `yaml:"port"`
}

var defaultWebServerAppConf *WebServerAppConfig

func GetDefaultWebServerAppConfig() *WebServerAppConfig {
	return defaultWebServerAppConf
}
