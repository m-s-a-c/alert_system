package main

import (
	"flag"
	"fmt"
	"log"

	// "slices"
	// "strconv"

	"github.com/m-s-a-c/alert_system.git/core/config"
	logging "github.com/m-s-a-c/alert_system.git/core/logger"
	chain "github.com/m-s-a-c/alert_system.git/core/providers/0chain"
	"github.com/m-s-a-c/alert_system.git/core/slack"
	"github.com/spf13/viper"
)

func initializeConfig() {
	config.Configuration.NetworkDomain = viper.GetString("network.domain")
	config.Configuration.NetworkSubdomain = viper.GetString("network.subdomain")

	config.Configuration.Port = viper.GetInt("alert.port")

	config.Configuration.UseHTTPS = viper.GetBool("use_https")
	config.Configuration.UsePath = viper.GetBool("use_path")

	config.Configuration.SlackWebhook = viper.GetString("slack.webhook")

	config.Configuration.KafkaEnabled = viper.GetBool("kafka.enabled")
	config.Configuration.KafkaHost = viper.GetString("kafka.host")
	config.Configuration.KafkaUsername = viper.GetString("kafka.username")
	config.Configuration.KafkaPassword = viper.GetString("kafka.password")
	config.Configuration.KafkaEventsTopic = viper.GetString("kafka.eventsTopic")
	config.Configuration.KafkaEventsGroupID = viper.GetString("kafka.eventsGroupId")
	config.Configuration.KafkaEventsRetryTopic = viper.GetString("kafka.eventsRetryTopic")

	config.Configuration.RedisHost = viper.GetString("redis.host")
	config.Configuration.RedisPassword = viper.GetString("redis.password")
	config.Configuration.RedisDB = viper.GetString("redis.db")

	config.Configuration.VaultKey = viper.GetString("vault.key")
	config.Configuration.PortainerHost = viper.GetString("portainer.host")
	config.Configuration.PortainerUsername = viper.GetString("portainer.username")
	config.Configuration.PortainerPassword = viper.GetString("portainer.password")

	config.Configuration.ElasticUsername = viper.GetString("elastic.username")
	config.Configuration.ElasticPassword = viper.GetString("elastic.password")
	config.Configuration.ElasticHost = viper.GetString("elastic.host")
	config.Configuration.ElasticPort = viper.GetString("elastic.port")
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

	//for key, value := range viper.AllSettings() {
	//	fmt.Printf("%s: %v\n", key, value)
	//}

	var blockWorkerUrl = "https://" + config.Configuration.NetworkSubdomain + "." + config.Configuration.NetworkDomain + "/network"
	fmt.Println("blockWorker url:", blockWorkerUrl)
	prov := chain.GetProvidersURL(blockWorkerUrl)
	activeSharders, inActiveSharders := chain.CheckUrlStatus(prov.Sharders)
	activeMiners, inActiveMiners := chain.CheckUrlStatus(prov.Miners)

	roundMiner, err := chain.CheckProviderRound(activeMiners)
	if err != nil {
		fmt.Println(err)
	}
	roundSharder, err := chain.CheckProviderRound(activeSharders)
	if err != nil {
		fmt.Println(err)
	}

	laggingMiners := chain.LaggingProviders(roundMiner)
	laggingSharders := chain.LaggingProviders(roundSharder)

	err = slack.SendSlackMessage(inActiveSharders, "unreachable", "SHARDERS", config.Configuration.SlackWebhook, "#testing")
	if err != nil {
		log.Fatalf("Slack notification failed: ", err)
	}
	err = slack.SendSlackMessage(inActiveMiners, "unreachable", "MINERS", config.Configuration.SlackWebhook, "#testing")
	if err != nil {
		log.Fatalf("Slack notification failed: ", err)
	}
	err = slack.SendSlackMessage(laggingMiners, "lagging", "MINERS", config.Configuration.SlackWebhook, "#testing")
	if err != nil {
		log.Fatalf("Slack notification failed: ", err)
	}
	err = slack.SendSlackMessage(laggingSharders, "lagging", "MINERS", config.Configuration.SlackWebhook, "#testing")
	if err != nil {
		log.Fatalf("Slack notification failed: ", err)
	}
}
