package configparsers

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
)

var config Config

func ConfigManager(path string) (Config,error) {
	log.Println("********************Prepairing the config with "+path+"**********************")
	parsedConfig, err := ioutil.ReadFile(path)
	if err != nil {
		return config,errors.New("Not able to read file "+ path)
	}
	return GetConfig(string(parsedConfig))
}

func GetConfig(apiConfig string) (Config,error) {
	var err error
		jsonErr := json.Unmarshal([]byte(apiConfig), &config)
		if jsonErr != nil {
			log.Println("Config error")
			err = errors.New("json format error: please check your json load profile")
		}
	return config, err
}