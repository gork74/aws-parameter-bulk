package conf

import (
	"github.com/spf13/viper"
)

// BindEnv binds used environment variables
func BindEnv() {
	viper.SetEnvPrefix("")
	viper.BindEnv("SSM_LOG_LEVEL")
}
