package main

import (
	// _ "net/http/pprof"

	"flag"
	"fmt"
	"goranger/actions"
	pool "goranger/concurrency"
	"goranger/config"
	parser "goranger/configparsers"
	utils "goranger/genericUtilities"
	"log"
	"os"

	//"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/tidwall/gjson"
	"sync"
	"time"
)

var Env parser.Environment
var PoolObjectMap = make(map[string]*pool.Pool)
var ActiveTaskMap = make(map[string]*pool.Pool)
var TaskObjectMap = make(map[string]parser.Task)
var TaskUpdateChannel <-chan time.Time
var TaskDeleteChannel = make(chan string, 100)
var IsTriggered = make(chan bool)
var wg sync.WaitGroup

func main() {

	flag.Set("logtostderr", "true")
	profile := flag.String("profile", "sanity", "name of json file present in profiles folder")
	flag.Parse()
	file, err := os.Create("stdout.log")
	log.SetOutput(file)
	config.GetEnvironment()
	config.SetProfile(*profile)
	config.LoadPropabilities(config.GetPropabilityFile())

	go TaskUpdateExecutor()
	go ConcurrencyUpdateExecutor()
	go PropabilityUpdateExecutor()
	go ThinkTimeUpdateExecutor()

	configs, err := parser.ConfigManager(config.GetProfile())
	if err != nil {
		//panic(err.Error())
		fmt.Println(err.Error())
		os.Exit(100)
	}

	for _, value := range configs.Tasks {
		if value.Object.Isenabled && value.Object.Rate > 0 {
			taskCount := &parser.TaskCount{
				Success: 0,
				Error:   0,
			}
			value.Object.TaskCount = taskCount
			var waitGroup sync.WaitGroup
			value.SetRedisConnectionPool()
			value.CreateGRPCConnection()
			csvErr := value.LoadCSVData()
			if csvErr != nil {
				fmt.Println(csvErr.Error())
				os.Exit(100)
			}
			TaskObjectMap[value.Name] = value
			var ratelimiter = utils.GetRateLimiter(int64(value.Object.Rate))
			genChan := make(chan string, 20000)
			doneChan := make(chan bool)
			destroyChan := make(chan bool, 20000)
			prodDestroyChan := make(chan bool, 20000)
			concurrency := value.GetConcurrency()
			p := pool.CreateNewPool(value, concurrency, value.Object.Duration, ratelimiter, genChan, doneChan, &waitGroup, destroyChan, prodDestroyChan)
			PoolObjectMap[value.Name] = p
			if config.IsAutoCreateFunctionsEnabled() {
				p.SetAction(actions.BaseFunc)
			}
			go p.Run(&wg, TaskDeleteChannel, IsTriggered)
			<-IsTriggered
		}
	}
	wg.Wait()
}

func TaskUpdateExecutor() {
	watcher, _ := fsnotify.NewWatcher()
	//TaskUpdateChannel = time.Tick(time.Second * time.Duration(config.GetTaskUpdateCheckInterval()))

	for {
		watcher.Add(config.GetProfile())
		<-watcher.Events
		//<- TaskUpdateChannel
		for key, value := range TaskObjectMap {
			if TaskObjectMap[key].Object.Isenabled {
				fileString, _ := utils.ReadFile(config.GetProfile())
				newRate := gjson.Get(fileString, `tasks.#[name="`+key+`"].taskobject.rate`).Int()
				if value.Object.Rate != int(newRate) {
					currentProducerCount := pool.GetTaskGeneratorSpawnCount(PoolObjectMap[key])
					ratelimiter := utils.GetRateLimiter(int64(newRate))
					PoolObjectMap[key].SetRateLimiter(ratelimiter)
					PoolObjectMap[key].SetTaskRate(int(newRate))
					expectedProducerCount := pool.GetTaskGeneratorSpawnCount(PoolObjectMap[key])
					producerCountDiff := expectedProducerCount - currentProducerCount

					if producerCountDiff > 0 {
						isTriggered := make(chan bool)
						for i := 0; i < producerCountDiff; i++ {
							go pool.TaskGenerator(PoolObjectMap[key], &wg, isTriggered)
							<-isTriggered
						}
					} else if producerCountDiff < 0 {
						for i := 0; i < expectedProducerCount; i++ {
							PoolObjectMap[key].ProducerDestroyChannel <- false
						}
						for i := 0; i < -(producerCountDiff); i++ {
							PoolObjectMap[key].ProducerDestroyChannel <- true
						}
					}

					value.Object.Rate = int(newRate)
					TaskObjectMap[key] = value
					value.SetRedisConnectionPool()
				}

			}
		}
	}
}

func ConcurrencyUpdateExecutor() {
	watcher, _ := fsnotify.NewWatcher()

	for {
		watcher.Add(config.GetProfile())
		<-watcher.Events
		for key, value := range TaskObjectMap {
			if TaskObjectMap[key].Object.Isenabled {
				fileString, _ := utils.ReadFile(config.GetProfile())
				newConcurrency := gjson.Get(fileString, `tasks.#[name="`+key+`"].taskobject.concurrency`).Int()
				PoolObjectMap[key].SetConcurrency(int(newConcurrency))

				if value.GetConcurrency() != int(newConcurrency) {

					concurrency := value.GetConcurrency()
					concDiff := int(newConcurrency) - concurrency

					if concDiff > 0 {
						for i := 0; i < concDiff; i++ {
							go pool.Stricker(PoolObjectMap[key])
						}
					} else if concDiff < 0 {
						for i := 0; i < int(concDiff); i++ {
							PoolObjectMap[key].DestroyChannel <- false
						}
						for i := 0; i < -(concDiff); i++ {
							PoolObjectMap[key].DestroyChannel <- true
						}
					}

					value.Object.Concurrency = int(newConcurrency)
					TaskObjectMap[key] = value
					value.SetRedisConnectionPool()
				}

			}
		}
	}
}

func ThinkTimeUpdateExecutor() {
	watcher, _ := fsnotify.NewWatcher()

	for {
		watcher.Add(config.GetProfile())
		<-watcher.Events
		for key, value := range TaskObjectMap {
			if TaskObjectMap[key].Object.Isenabled {
				fileString, _ := utils.ReadFile(config.GetProfile())
				newThinkTime := gjson.Get(fileString, `tasks.#[name="`+key+`"].taskobject.thinkTime`).String()
				PoolObjectMap[key].SetThinkTime(newThinkTime)
				preMin, preMax, _ := value.GetThinkTime()
				value.Object.ThinkTime = newThinkTime
				TaskObjectMap[key] = value
				curMin, curMax, _ := value.GetThinkTime()

				if preMin != curMin || preMax != curMax {
					PoolObjectMap[key].SetMinThinkTime(curMin)
					PoolObjectMap[key].SetMaxThinkTime(curMax)
				}

			}
		}
	}
}

func PropabilityUpdateExecutor() {
	for {
		watcher, _ := fsnotify.NewWatcher()
		watcher.Add(config.GetPropabilityFile())
		<-watcher.Events
		config.LoadPropabilities(config.GetPropabilityFile())
	}
}
