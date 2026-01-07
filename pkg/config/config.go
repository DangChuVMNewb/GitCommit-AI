package config

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
)

type Config struct {
	ApiKey   string `json:"api_key"`
	Language string `json:"language"`
}

func GetConfigDir() string {
	var configDir string
	if runtime.GOOS == "windows" {
		configDir = os.Getenv("APPDATA")
		if configDir == "" {
			configDir = filepath.Join(os.Getenv("USERPROFILE"), "AppData", "Roaming")
		}
	} else {
		home, _ := os.UserHomeDir()
		configDir = filepath.Join(home, ".config")
	}
	return filepath.Join(configDir, "commitai")
}

func GetConfigFile() string {
	return filepath.Join(GetConfigDir(), "config.json")
}

func Load() Config {
	file := GetConfigFile()
	data, err := ioutil.ReadFile(file)
	var config Config
	if err == nil {
		json.Unmarshal(data, &config)
	}
	if config.Language == "" {
		config.Language = "en"
	}
	return config
}

func Save(config Config) error {
	dir := GetConfigDir()
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		os.MkdirAll(dir, 0755)
	}
	data, _ := json.MarshalIndent(config, "", "  ")
	return ioutil.WriteFile(GetConfigFile(), data, 0644)
}
