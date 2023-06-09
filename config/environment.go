package config

import (
	"encoding/json"
	"goranger/helpers"
	"io/ioutil"
	"strings"
	"sync"
)

var onceEnvironment sync.Once
var Env Environment
var JSON *helpers.JSON

type Environment struct {
	Debug                        bool   `json:"debug"`
	ExecutorMultiplierCount      int    `json:"executorMultiplierCount"`
	ProducersBaseCountPerForLoop int    `json:"producerBaseCountPerForLoop"`
	Profile                      string `json:"profile"`
	PoolDebug                    bool   `json:"poolDebug"`
	Prometheus                   bool   `json:"prometheus"`
	TaskUpdateCheckInterval      int    `json:"taskUpdateCheckInterval"`
	AutoCreateFunctions          bool   `json:"autoCreateFunctions"`
	PropabilityFile              string `json:"propabilityFile"`
}

func GetEnvironment() {
	onceEnvironment.Do(func() {
		file, err := ioutil.ReadFile("config/env.json")
		if err != nil {
			panic(err.Error())
		}
		json.Unmarshal(file, &Env)
	})

}

func SetProfile(profile string) {
	Env.Profile = profile
}

func IsDebug() bool {
	return Env.Debug
}

func IsSanity() bool {
	if Env.Profile == "sanity" {
		return true
	}
	return false
}

func IsPoolDebug() bool {
	return Env.PoolDebug
}

func GetProfile() string {
	if strings.Contains(Env.Profile, ".json") {
		return Env.Profile
	} else {
		return "profiles/" + Env.Profile + ".json"
	}

}

func GetExecutorMultiplierCount() int {
	return Env.ExecutorMultiplierCount
}

func GetProducersBaseCountPerForLoop() int {
	return Env.ProducersBaseCountPerForLoop
}

func GetTaskUpdateCheckInterval() int {
	return Env.TaskUpdateCheckInterval
}

func IsPrometheusEnabled() bool {
	return Env.Prometheus
}

func IsAutoCreateFunctionsEnabled() bool {
	return Env.AutoCreateFunctions
}

func GetPropabilityFile() string {
	return Env.PropabilityFile
}

func LoadPropabilities(filepath string) error {
	json, err := helpers.LoadJSON(filepath)
	JSON = json
	return err
}
