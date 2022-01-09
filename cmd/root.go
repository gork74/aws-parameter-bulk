package cmd

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/gork74/aws-parameter-bulk/conf"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/natefinch/lumberjack.v2"
)

// Config and global logger
var configFile string
var pidFile string
var logger zerolog.Logger

// The Root Cobra Handler
var rootCmd = &cobra.Command{
	Version: conf.Version,
	Use:     conf.Executable,
}

func main() {
	Execute()
}

// This is the main initializer handling cli, config and log
func init() { // nolint: gochecknoinits
	// Initialize configuration
	cobra.OnInitialize(conf.BindEnv, initConfig, initLog)
}

// Execute starts the program
func Execute() {
	// Run the program
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err.Error())
	}
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {

	// Sets up the config file, environment etc
	viper.SetEnvPrefix(strings.ToUpper(conf.Executable))
	// If a default value is []string{"a"} an environment variable of "a b" will end up []string{"a","b"}
	viper.SetTypeByDefaultValue(true)
	// Automatically use environment variables where available
	viper.AutomaticEnv()
	// Environment variables use underscores instead of periods
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	viper.Set("logger.level", viper.GetString("SSM_LOG_LEVEL"))

}

func initLog() {

	zerolog.TimeFieldFormat = zerolog.TimeFormatUnixMs

	// log level
	var logLevel zerolog.Level
	var err error
	logLevel, err = zerolog.ParseLevel(viper.GetString("logger.level"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to parse log level: %s ERROR: %s\n", viper.GetString("logger.level"), err.Error())
		logLevel = zerolog.DebugLevel
	}

	zerolog.SetGlobalLevel(logLevel)

	var logWriter io.Writer

	if viper.GetString("logger.file") != "" {
		logWriter = &lumberjack.Logger{
			Filename:   viper.GetString("logger.file"),
			MaxSize:    100, // megabytes
			MaxBackups: 3,
			MaxAge:     28, // days
		}
	} else {
		// log on stdout
		// pretty console logger
		output := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}

		logWriter = output
	}

	logger = zerolog.New(logWriter).With().Timestamp().Caller().Logger()

	// set global logger
	log.Logger = logger
}
