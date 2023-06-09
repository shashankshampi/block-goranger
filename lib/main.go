package lib

import (
	// _ "net/http/pprof"

	"errors"
	"flag"
	pool "goranger/concurrency"
	"goranger/config"
	parser "goranger/configparsers"
	utils "goranger/genericUtilities"
	"log"
	"os"
	"sync"
)

type Actions func(task parser.Task)

var taskChannel chan string
var doneTaskChannel chan bool
var deleteTaskChannel chan string
var hudor *Hudor

type Hudor struct {
	Env           config.Environment
	TaskObjectMap map[string]parser.Task
	PoolObjectMap map[string]*pool.Pool
	ActiveTaskMap map[string]*pool.Pool
	WG            *sync.WaitGroup
}

func createHudorObject(env config.Environment, taskObjectMap map[string]parser.Task, poolObjectMap map[string]*pool.Pool, activeTaskMap map[string]*pool.Pool, wg *sync.WaitGroup) *Hudor {
	return &Hudor{
		Env:           env,
		TaskObjectMap: taskObjectMap,
		PoolObjectMap: poolObjectMap,
		ActiveTaskMap: activeTaskMap,
		WG:            wg,
	}
}

func Init(loadProfile string) (*Hudor, error) {

	flag.Set("logtostderr", "true")
	flag.Parse()
	file, err := os.Create("stdout.log")
	log.SetOutput(file)

	if hudor != nil {
		hudor.Close()
	}

	var wg sync.WaitGroup
	PoolObjectMap := make(map[string]*pool.Pool)
	ActiveTaskMap := make(map[string]*pool.Pool)
	taskChannel = make(chan string, 100)
	deleteTaskChannel = make(chan string, 100)
	doneTaskChannel = make(chan bool, 100)

	config.GetEnvironment()
	configs, err := parser.ConfigManager(loadProfile)

	var taskObjMap = make(map[string]parser.Task)
	for _, value := range configs.Tasks {
		if value.Object.Isenabled && value.Object.Rate > 0 {
			var waitGroup sync.WaitGroup

			taskCount := &parser.TaskCount{
				Success: 0,
				Error:   0,
			}
			value.Object.TaskCount = taskCount

			value.SetRedisConnectionPool()
			value.CreateGRPCConnection()
			taskObjMap[value.Name] = value
			var ratelimiter = utils.GetRateLimiter(int64(value.Object.Rate))
			genChan := make(chan string, 20000)
			doneChan := make(chan bool)
			destroyChan := make(chan bool, 20000)
			prodDestroyChan := make(chan bool, 20000)
			p := pool.CreateNewPool(value, value.Object.Rate, value.Object.Duration, ratelimiter, genChan, doneChan, &waitGroup, destroyChan, prodDestroyChan)
			PoolObjectMap[value.Name] = p
		}
	}

	hudorObj := createHudorObject(config.Env, taskObjMap, PoolObjectMap, ActiveTaskMap, &wg)
	hudor = hudorObj
	go hudorObj.taskExecutor()
	go hudorObj.taskDelete()
	return hudorObj, err
}

func (hudor *Hudor) taskExecutor() {

	hudor.WG.Add(1)
	for {
		taskName := <-taskChannel
		if taskName != "" {
			poolObject := hudor.PoolObjectMap[taskName]
			hudor.ActiveTaskMap[taskName] = hudor.PoolObjectMap[taskName]
			isTriggered := make(chan bool, 100)
			go poolObject.Run(hudor.WG, deleteTaskChannel, isTriggered)
			<-isTriggered
			doneTaskChannel <- true
		} else {
			deleteTaskChannel <- ""
			break
		}
	}
	//fmt.Println("I M HERE")
	hudor.WG.Done()
}

func (hudor *Hudor) taskDelete() {
	for {
		taskName := <-deleteTaskChannel
		if taskName != "" {
			delete(hudor.ActiveTaskMap, taskName)
		} else {
			break
		}
	}

}

func (hudor *Hudor) Run(taskName string, actions Actions) error {

	_, isTaskPresent := hudor.TaskObjectMap[taskName]
	_, isRunning := hudor.ActiveTaskMap[taskName]

	if isTaskPresent && !isRunning {
		ratelimiter := utils.GetRateLimiter(int64(hudor.TaskObjectMap[taskName].Object.Rate))

		poolObject := hudor.PoolObjectMap[taskName]
		taskObjectMap := hudor.TaskObjectMap[taskName]
		csvErr := taskObjectMap.LoadCSVData()
		if csvErr != nil {
			return csvErr
		}
		taskObjectMap.Object.IsStopped = false
		taskObjectMap.Object.TaskCount.Success = 0
		taskObjectMap.Object.TaskCount.Error = 0
		hudor.TaskObjectMap[taskName] = taskObjectMap
		poolObject.StartTask()
		hudor.PoolObjectMap[taskName] = poolObject
		poolObject.SetRateLimiter(ratelimiter)
		poolObject.SetAction(actions)
		taskChannel <- taskName
		<-doneTaskChannel
	} else if !isTaskPresent {
		return errors.New("Task not present in the profile json")
	} else if isRunning {
		return errors.New("Task already in running state")
	}
	return nil
}

func (hudor *Hudor) UpdateTaskRate(taskName string, rate int) error {
	_, isRunning := hudor.ActiveTaskMap[taskName]
	if isRunning {

		taskObjectMap := hudor.TaskObjectMap[taskName]
		currentProducerCount := pool.GetTaskGeneratorSpawnCount(hudor.PoolObjectMap[taskName])
		oldRate := taskObjectMap.Object.Rate
		ratelimiter := utils.GetRateLimiter(int64(rate))
		hudor.PoolObjectMap[taskName].SetRateLimiter(ratelimiter)
		hudor.PoolObjectMap[taskName].SetTaskRate(int(rate))
		expectedProducerCount := pool.GetTaskGeneratorSpawnCount(hudor.PoolObjectMap[taskName])
		producerCountDiff := expectedProducerCount - currentProducerCount
		rateDiff := int(rate) - oldRate

		if producerCountDiff > 0 {
			isTriggered := make(chan bool)
			for i := 0; i < producerCountDiff; i++ {
				go pool.TaskGenerator(hudor.PoolObjectMap[taskName], hudor.WG, isTriggered)
				<-isTriggered
			}
		} else if producerCountDiff < 0 {
			for i := 0; i < expectedProducerCount; i++ {
				hudor.PoolObjectMap[taskName].ProducerDestroyChannel <- false
			}
			for i := 0; i < -(producerCountDiff); i++ {
				hudor.PoolObjectMap[taskName].ProducerDestroyChannel <- true
			}
		}

		if rateDiff > 0 {
			for i := 0; i < rateDiff; i++ {
				go pool.Stricker(hudor.PoolObjectMap[taskName])
			}
		} else if rateDiff < 0 {
			for i := 0; i < int(rate); i++ {
				hudor.PoolObjectMap[taskName].DestroyChannel <- false
			}
			for i := 0; i < -(rateDiff); i++ {
				hudor.PoolObjectMap[taskName].DestroyChannel <- true
			}
		}
		taskObjectMap.Object.Rate = rate
		hudor.TaskObjectMap[taskName] = taskObjectMap
		taskObjectMap.SetRedisConnectionPool()
	} else {
		return errors.New(taskName + " not in running state")
	}

	return nil
}

func (hudor *Hudor) WaitForTaskToComplete(taskName string) error {

	_, isTaskPresent := hudor.ActiveTaskMap[taskName]

	if isTaskPresent {
		poolObject := hudor.PoolObjectMap[taskName]
		poolObject.GetWaitGroup().Wait()
		return nil
	} else {
		return errors.New(taskName + " not in running state")
	}
}

func (hudor *Hudor) StopTask(taskName string) error {
	_, isTaskPresent := hudor.ActiveTaskMap[taskName]

	if isTaskPresent {
		poolObject := hudor.PoolObjectMap[taskName]
		poolObject.StopTask()
		taskObjectMap := hudor.TaskObjectMap[taskName]
		taskObjectMap.Object.IsStopped = true
		hudor.TaskObjectMap[taskName] = taskObjectMap
		poolObject.GetWaitGroup().Wait()
		return nil
	} else {
		return errors.New(taskName + " not in running state")
	}
}

func (hudor *Hudor) StopAllTasks() {
	for key, _ := range hudor.ActiveTaskMap {
		hudor.StopTask(key)
	}
}

func (hudor *Hudor) WaitForAllTasksToComplete() {
	for key, _ := range hudor.PoolObjectMap {
		hudor.WaitForTaskToComplete(key)
	}
}

func (hudor *Hudor) Close() {
	taskChannel <- ""
	hudor.WG.Wait()
}
