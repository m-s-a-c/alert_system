package main

import (
	"flag"
	"fmt"

	"github.com/m-s-a-c/alert_system.git/core/config"
	logging "github.com/m-s-a-c/alert_system.git/core/logger"
	"github.com/spf13/viper"
)

func initializeConfig() {
	config.Configuration.NetworkDomain = viper.GetString("network.domain")
	config.Configuration.NetworkSubdomain = viper.GetString("network.subdomain")
	config.Configuration.Port = viper.GetInt("alert.port")
	config.Configuration.UseHTTPS = viper.GetBool("use_https")
	config.Configuration.UsePath = viper.GetBool("use_path")
}

func main() {
	deploymentMode := flag.Int("deployment_mode", 0, "Deployment mode")
	configDir := flag.String("config_dir", "./docker.local/config", "Configuration Directory")
	logDir := flag.String("log_dir", "", "log_dir")
	flag.Parse()

	config.Configuration.DeploymentMode = byte(*deploymentMode)
	config.SetupConfig(*configDir)

	if config.DevNet() {
		logging.InitLogging("development", *logDir)
	} else {
		logging.InitLogging("production", *logDir)
	}

	initializeConfig()

	for key, value := range viper.AllSettings() {
		fmt.Printf("%s: %v\n", key, value)
	}
}
