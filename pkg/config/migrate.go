package config

// CurrentConfigVersion is the latest supported config schema version.
const CurrentConfigVersion = 1

// MigrateConfig checks the config version and applies any needed migrations.
// Returns (migrated bool, summary string, err error).
func MigrateConfig(cfg *ConfigData) (bool, string, error) {
	if cfg.Version >= CurrentConfigVersion {
		return false, "already at latest version", nil
	}
	// Version 0 (unset) → 1: just set the version field.
	cfg.Version = CurrentConfigVersion
	return true, "migrated from version 0 to version 1 (added version field)", nil
}
