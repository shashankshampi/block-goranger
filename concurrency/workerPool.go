package concurrency

import (
	"goranger/actions"
	"goranger/config"
	"goranger/constants"
	"goranger/genericUtilities"
	"math"
	"sync"
	"sync/atomic"
	"time"

	parser "goranger/configparsers"
)

type Pool struct {
	taskObj                parser.Task
	concurrency            int
	ratelimiter            <-chan time.Time
	rampUpTime             int
	minThinkTime           int
	maxThinkTime           int
	futureTimeInMin        int
	channel                chan string
	donechannel            chan bool
	waitGroup              *sync.WaitGroup
	actions                func(task parser.Task)
	DestroyChannel         chan bool
	ProducerDestroyChannel chan bool
	metrics                Metrics
}

type Metrics struct {
	noOfExecutors     int64
	noOfProducers     int64
	producerCount     int64
	executorsCount    int64
	noOfRequestQueued int64
}

func CreateNewPool(taskObj parser.Task, poolCount int, futureTimeInMin int, ratelimiter <-chan time.Time, genChan chan string, doneChan chan bool, wg *sync.WaitGroup, destroyChan chan bool, prodDestroyChan chan bool) *Pool {

	min, max, _ := taskObj.GetThinkTime()
	return &Pool{
		taskObj:                taskObj,
		concurrency:            poolCount,
		ratelimiter:            ratelimiter,
		rampUpTime:             taskObj.GetRampUpTime(),
		minThinkTime:           min,
		maxThinkTime:           max,
		futureTimeInMin:        futureTimeInMin,
		channel:                genChan,
		donechannel:            doneChan,
		waitGroup:              wg,
		actions:                nil,
		DestroyChannel:         destroyChan,
		ProducerDestroyChannel: prodDestroyChan,

		metrics: Metrics{
			noOfExecutors: 0,
			noOfProducers: 0,
		},
	}
}

func (p *Pool) SetAction(action func(task parser.Task)) {
	p.actions = action
}

func (p *Pool) Run(wg *sync.WaitGroup, deleteTaskChannel chan string, isTriggered chan bool) {
	spawntaskGenerator(p, wg)
	spawnRoutines(p, wg, deleteTaskChannel, isTriggered)
}

func (p *Pool) GetWaitGroup() *sync.WaitGroup {
	return p.waitGroup
}

func (p *Pool) GetDoneChannel() chan bool {
	return p.donechannel
}

func (p *Pool) SetRateLimiter(ratelimiter <-chan time.Time) {
	p.ratelimiter = ratelimiter
}

func (p *Pool) SetTaskRate(rate int) {
	p.taskObj.Object.Rate = rate
}

func (p *Pool) SetConcurrency(concurrency int) {
	p.taskObj.Object.Concurrency = concurrency
}

func (p *Pool) SetThinkTime(thinkTime string) {
	p.taskObj.Object.ThinkTime = thinkTime
}

func (p *Pool) SetMinThinkTime(min int) {
	p.minThinkTime = min
}

func (p *Pool) SetMaxThinkTime(max int) {
	p.maxThinkTime = max
}

func (p *Pool) StopTask() {
	p.taskObj.Object.IsStopped = true
}

func (p *Pool) StartTask() {
	p.taskObj.Object.IsStopped = false
}

func (p *Pool) GetStus() bool {
	return p.taskObj.Object.IsStopped
}

func spawnRoutines(p *Pool, wg *sync.WaitGroup, deleteTaskChannel chan string, isTriggered chan bool) {

	for i := 0; i < p.concurrency; i++ {
		go Stricker(p)
		atomic.AddInt64(&p.metrics.noOfExecutors, 1)
		if p.rampUpTime != 0 {
			time.Sleep(time.Second * time.Duration(p.rampUpTime))
		}
	}
	count := 0
	for {
		isTriggered <- true
		if count >= GetTaskGeneratorSpawnCount(p) {
			deleteTaskChannel <- p.taskObj.Name
			break
		}
		<-p.donechannel
		p.waitGroup.Done()
		wg.Done()
		count++
	}
}

func GetTaskGeneratorSpawnCount(p *Pool) int {
	return int(math.Ceil(float64(p.taskObj.Object.Rate) / float64(config.GetProducersBaseCountPerForLoop())))
}

func spawntaskGenerator(p *Pool, wg *sync.WaitGroup) {
	isTriggered := make(chan bool)
	for i := 0; i < GetTaskGeneratorSpawnCount(p); i++ {
		go TaskGenerator(p, wg, isTriggered)
		atomic.AddInt64(&p.metrics.noOfProducers, 1)
		<-isTriggered
	}
}

func TaskGenerator(p *Pool, wg *sync.WaitGroup, isTriggered chan bool) {
	wg.Add(1)
	p.waitGroup.Add(1)
	var now = time.Now()
	var future = time.Now().Add(time.Minute * time.Duration(p.futureTimeInMin))
	go func(p *Pool) {
		rateLimiter := p.ratelimiter
		currentProducerCount := GetTaskGeneratorSpawnCount(p)
		isDestroy := false
		for now.Unix() < future.Unix() {
			now = time.Now()
			<-p.ratelimiter
			p.channel <- "1"
			genericUtilities.Increment(p.taskObj.Name + constants.PRODUCER_COUNT)
			atomic.AddInt64(&p.metrics.producerCount, 1)
			if p.taskObj.Object.IsStopped {
				break
			}
			if p.ratelimiter != rateLimiter {
				if GetTaskGeneratorSpawnCount(p) < currentProducerCount {
					isDestroy = <-p.ProducerDestroyChannel
					if isDestroy {
						atomic.AddInt64(&p.metrics.noOfProducers, -1)
						wg.Done()
						p.waitGroup.Done()
						break
					}
				}
				rateLimiter = p.ratelimiter
				currentProducerCount = GetTaskGeneratorSpawnCount(p)
			}
		}
		if !isDestroy {
			p.donechannel <- true
		}
		genericUtilities.PushLatency(p.taskObj.Name+constants.PRODUCER_LATENCY, now)
	}(p)
	isTriggered <- true
}

func Stricker(p *Pool) {
	concurrency := p.concurrency
	currentConcurrency := p.taskObj.Object.Concurrency
	isDestroy := false
	executorChannel := make(chan bool)
	for {
		now := time.Now()
		<-p.channel
		genericUtilities.Increment(p.taskObj.Name + constants.EXECUTOR_COUNT)
		if p.actions != nil {
			go libExecutor(p, executorChannel)
		} else {
			go executor(p, executorChannel)
		}
		if p.taskObj.Object.Sync {
			<-executorChannel
		}
		if p.taskObj.Object.IsStopped {
			break
		}
		if p.concurrency != concurrency {
			if p.taskObj.Object.Concurrency < currentConcurrency {
				isDestroy = <-p.DestroyChannel
				if isDestroy {
					atomic.AddInt64(&p.metrics.noOfExecutors, -1)
					break
				}
			}
			concurrency = p.concurrency
			currentConcurrency = p.taskObj.Object.Rate
		}
		if p.minThinkTime != 0 && p.maxThinkTime == 0 {
			time.Sleep(time.Duration(p.minThinkTime) * time.Millisecond)
		} else if p.maxThinkTime != 0 {
			duration := genericUtilities.GetRandomValue(p.maxThinkTime) + p.minThinkTime
			time.Sleep(time.Duration(duration) * time.Millisecond)
		}
		genericUtilities.PushLatency(p.taskObj.Name+constants.EXECUTOR_LATENCY, now)
	}
}

func libExecutor(p *Pool, executorChan chan bool) {
	p.actions(p.taskObj)
	if p.taskObj.Object.Sync {
		executorChan <- true
	}

}

func executor(p *Pool, executorChan chan bool) {
	now := time.Now()
	genericUtilities.Increment(p.taskObj.Name + constants.REQUEST_COUNT)
	switch p.taskObj.Name {
	//add case for your actions here
	case "SampleRequest":
		actions.SampleRequest(p.taskObj)
	}
	genericUtilities.PushLatency(p.taskObj.Name+constants.REQUEST_LATENCY, now)
	if p.taskObj.Object.Sync {
		executorChan <- true
	}
}
