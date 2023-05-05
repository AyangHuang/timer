package conf

type MySQLConfig struct {
	DSN          string `yaml:"dsn"`
	MaxOpenConns int    `yaml:"maxOpenConns"`
	MaxIdleConns int    `yaml:"maxIdleConns"`
}

var defaultMySQLConfig *MySQLConfig

func GetDefaultMySQLConfig() *MySQLConfig {
	return defaultMySQLConfig
}
