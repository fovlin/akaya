package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path"

	"github.com/fovlin/record"
)

const (
	configFile string = "akaya.json"
)

type DirEntry struct {
	Type string
	Time string
	Size string
}

type Config struct {
	Ip     string  `json:"ip"`
	Port   float64 `json:"port"`
	Root string `json:"root"`
	TLS TLSConfig `json:"tls"`
}

type TLSConfig struct {
	Enable    bool    `json:"enable"`
	CRT string  `json:"crt"`
	Key string  `json:"key"`
}

var (
	configFilePath string
	config Config
	defaultConfig Config = Config{
		Ip: "0.0.0.0",
		Port: 80,
		Root: "./",
		TLS: TLSConfig{
			Enable: false,
			CRT: "",
			Key: "",
		},
	}

	specURL map[string]func(http.ResponseWriter, *http.Request) = map[string]func(http.ResponseWriter, *http.Request){

	}
)

func main() {

	if err := initConfigFilePath(); err != nil {
		record.Error("check config file:",err)
	}

	if err := loadConfig(); err != nil {
		record.Error("load config:", err)
	}

	var handler handler

	if config.TLS.Enable {
		record.Info("https server start on: https://" + config.Ip + ":" + fmt.Sprint(int(config.Port)))
		if err := http.ListenAndServeTLS(config.Ip+":"+fmt.Sprint(int(config.Port)), config.TLS.CRT, config.TLS.Key, handler); err != nil {
			record.Error("start https server", err)
		}
	} else {
		record.Info("http server start on: http://" + config.Ip + ":" + fmt.Sprint(int(config.Port)))
		if err := http.ListenAndServe(config.Ip+":"+fmt.Sprint(int(config.Port)), handler); err != nil {
			record.Error("start https server:", err)
		}
	}
}

func loadConfig() (error) {

	userConfigDir, err := os.UserConfigDir()
	if err != nil {
		return err
	}

	configFilePath = path.Join(userConfigDir, configFile)

	_, err = os.Stat(configFilePath)
	if os.IsNotExist(err) {
		if err = createConfig(); err != nil {
			record.Error(err)
		}
	} else if !os.IsNotExist(err) && err != nil {
		return err
	}

	configJsonData, err := os.ReadFile(configFilePath)
	if err != nil {
		return err
	}

	err = json.Unmarshal(configJsonData, &config)
	if err != nil {
		return err
	}

	return nil
}

func createConfig() (error) {

	file, err := os.Create(path.Join(configFilePath))
	if err != nil {
		return err
	}

	encoder := json.NewEncoder(file)
	encoder.SetIndent("","  ")
	err = encoder.Encode(defaultConfig)

	return nil

}

func initConfigFilePath() error {

	userConfigDir, err := os.UserConfigDir()
	if err != nil {
		return nil
	}

	if _, err := os.Stat(userConfigDir); os.IsNotExist(err) {
		if err := os.MkdirAll(userConfigDir, 0711); err != nil {
			return err
		}
	}


	configFilePath = path.Join(userConfigDir, configFile)

	return nil
}