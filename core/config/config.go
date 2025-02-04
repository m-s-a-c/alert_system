package config

import (
	"log"
	"strings"
	"sync"

	"github.com/spf13/viper"
)

const (
	DeploymentDev  = 0
	DeploymentProd = 1
)

// Config holds all configuration options passed from command line
type Config struct {
	NetworkSubdomain string
	NetworkDomain    string
	Port             int
	DeploymentMode   byte
	UseHTTPS         bool
	UsePath          bool
}

var (
	Configuration Config
	once          sync.Once
)

func SetupConfig(configDir string) {
	once.Do(func() {
		viper.SetDefault("logging.default", "info")
		viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
		viper.AutomaticEnv()
		viper.SetConfigName("alert")

		configPath := "./config"
		if configDir != "" {
			configPath = configDir
		}
		viper.AddConfigPath(configPath)

		if err := viper.ReadInConfig(); err != nil {
			log.Fatalf("Error reading config file: %v", err)
		}
	})
}

func DevNet() bool {
	return Configuration.DeploymentMode == DeploymentDev
}
