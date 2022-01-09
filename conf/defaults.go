package conf

import (
	"github.com/spf13/viper"
)

// Initialize defaults
var (
	_ = func() struct{} {
		// Logger Defaults
		viper.SetDefault("logger.level", "info")
		// if no file is specified, log on standard output
		viper.SetDefault("logger.file", "")

		viper.SetDefault("SSM_LOG_LEVEL", "info")

		return struct{}{}
	}()
)
