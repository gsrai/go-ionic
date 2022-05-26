package config

import (
	"encoding/json"
	"flag"
	"log"
	"os"
)

const devConfigFile = "dev.config.json"
const prodConfigFile = "prod.config.json"

const DEFAULT_HOST = "127.0.0.1"
const DEFAULT_PORT = "8080"

type Config struct {
	InputFilePath string `json:"INPUT_FILE_PATH"`
	ServerHost    string `json:"HOST"`
	ServerPort    string `json:"PORT"`
	CovalentAPI   struct {
		URL   string `json:"URL"`
		Token string `json:"KEY"`
	} `json:"COVALENT_API"`
	EtherscanAPI struct {
		URL   string `json:"URL"`
		Token string `json:"KEY"`
	} `json:"ETHERSCAN_API"`
	initialised bool
}

var configuration Config

func Init() {
	var configFile string
	isDev := flag.Bool("dev", true, "run application in development mode")
	flag.Parse()

	if *isDev {
		log.Println("running in DEVELOPMENT mode")
		configFile = devConfigFile

	} else {
		log.Println("running in PRODUCTION mode")
		configFile = prodConfigFile
	}

	data, err := os.ReadFile(configFile)
	if err != nil {
		log.Fatalf("Could not find or open config file: %s, %v\n", configFile, err)
	}

	cfg := Config{ServerHost: DEFAULT_HOST, ServerPort: DEFAULT_PORT}
	json.Unmarshal(data, &cfg)
	cfg.initialised = true
	configuration = cfg
}

func Get() Config {
	if !configuration.initialised {
		log.Fatal("Configuration is not initialised, make sure to call config.Init() in main")
	}
	return configuration
}
