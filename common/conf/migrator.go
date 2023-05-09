package conf

type MigratorAppConfig struct {
	MigrateStepMinutes          int `yaml:"migrateStepMinutes"`
	MigrateSuccessExpireMinutes int `yaml:"migrateSuccessExpireMinutes"`
	MigrateTryLockMinutes       int `yaml:"migrateTryLockMinutes"`
	TimerDetailCacheMinutes     int `yaml:"timerDetailCacheMinutes"`
}

var defaultMigratorAppConfig *MigratorAppConfig

func GetDefaultMigratorAppConfig() *MigratorAppConfig {
	return defaultMigratorAppConfig
}
