package conf

type MigratorAppConfig struct {
	WorkersNum                  int `yaml:"workersNum"`
	MigrateStepMinutes          int `yaml:"migrateStepMinutes"`
	MigrateSuccessExpireMinutes int `yaml:"migrateSuccessExpireMinutes"`
	MigrateTryLockMinutes       int `yaml:"migrateTryLockMinutes"`
	TimerDetailCacheMinutes     int `yaml:"timerDetailCacheMinutes"`
}

var defaultMigratorAppConfig *MigratorAppConfig

func GetDefaultMigratorAppConfig() *MigratorAppConfig {
	return defaultMigratorAppConfig
}
