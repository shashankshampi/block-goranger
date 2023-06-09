package configparsers

import (
	"context"
	"errors"
	"github.com/gomodule/redigo/redis"
	"google.golang.org/grpc"
	"goranger/constants"
	"goranger/csv"
	"goranger/grpcHelpers"
	redisUtility "goranger/redis"
	"strconv"
	"strings"
	"sync"
)

//type Config struct {
//	Tasks []struct {
//		Name       string `json:"name"`
//		Object     configparsers.TaskObject `json:"taskobject"`
//}

var mutex sync.Mutex

type BaseActions struct {
}

type Config struct {
	Tasks []Task `json:"tasks"`
}

type TaskCount struct {
	Success int
	Error   int
}

func (taskCount *TaskCount) IncrementSuccess() {
	mutex.Lock()
	taskCount.Success++
	mutex.Unlock()
}

func (taskCount *TaskCount) IncrementError() {
	mutex.Lock()
	taskCount.Error++
	mutex.Unlock()
}

func (taskCount *TaskCount) GetSuccessCount() int {
	return taskCount.Success
}

func (taskCount *TaskCount) GetErrorCount() int {
	return taskCount.Error
}

type Task struct {
	Name   string     `json:"name"`
	Object TaskObject `json:"taskobject"`
}

type Environment struct {
	Debug     bool `json:"debug"`
	Sanity    bool `json:"sanity"`
	PoolCount int  `json:"poolCount"`
}

type TaskObject struct {
	Protocol          string       `json:"protocol"`
	Rate              int          `json:"rate"`
	Concurrency       int          `json:"concurrency,omitempty"`
	RampUpTime        int          `json:"rampUpTime,omitempty"`
	Sync              bool         `json:"sync,omitempty"`
	ThinkTime         string       `json:"thinkTime,omitempty"`
	Duration          int          `json:"duration"`
	Httpendpoint      Httpendpoint `json:"httpendpoint,omitempty"`
	Grpcendpoint      Grpcendpoint `json:"grpcendpoint,omitempty"`
	File              []File       `json:"file,omitempty"`
	Redis             []Redis      `json:"redis,omitempty"`
	Isenabled         bool         `json:"isenabled"`
	IsStopped         bool         `json:"isstopped,omitempty"`
	ResponseExtractor []string     `json:"responseExtractor,omitempty"`
	TaskCount         *TaskCount   `json:"taskCount,omitempty"`
}

type File struct {
	Name           string `json:"name"`
	Path           string `json:"path"`
	LineSeperator  string `json:"lineseperator"`
	ValueSeperator string `json:"valueseperator"`
	IsSerialRead   bool   `json:"isSerial,omitempty"`
	IsEOF          bool   `json:"isEOF,omitempty"`
}

type Httpendpoint struct {
	URL         string   `json:"url"`
	Method      string   `json:"method"`
	PayloadPath string   `json:"payloadPath"`
	Headers     []string `json:"headers"`
	Cookies     struct {
	} `json:"cookies"`
}

type Grpcendpoint struct {
	Host        string           `json:"host"`
	Port        string           `json:"port"`
	APIPath     string           `json:"apiPath,omitempty"`
	PayloadPath string           `json:"payloadPath,omitempty"`
	Headers     []string         `json:"headers,omitempty"`
	Dialoptions []string         `json:"dialoptions,omitempty"`
	Connection  *grpc.ClientConn `json:"connection,omitempty"`
	Context     context.Context  `json:"context,omitempty"`
}

func (task *Task) CreateGRPCConnection() {
	if task.Object.Protocol == "grpc" {
		conn, ctx := grpcHelpers.GetGRPCConnection(task.Object.Grpcendpoint.Host, task.Object.Grpcendpoint.Port, task.Object.Grpcendpoint.Dialoptions)
		task.Object.Grpcendpoint.Connection = conn
		task.Object.Grpcendpoint.Context = ctx
	}
}

type PreRequisites struct {
	Tasks Config
}

type Redis struct {
	Name       string      `json:"name"`
	Host       string      `json:"host"`
	Password   string      `json:"password"`
	Index      int         `json:"index"`
	Rediskeys  []string    `json:"rediskeys"`
	Connection *redis.Pool `json:"connection,omitempty"`
}

func (r *Redis) SetConnection(poolObject *redis.Pool) {
	r.Connection = poolObject
}

func (task *Task) SetRedisConnectionPool() {

	for index, redisData := range task.Object.GetRedis() {
		task.Object.GetRedis()[index].Connection = redisUtility.GetRedisConnPool(task.Object.Rate, redisData.GetRedisConnectionUrl(), redisData.GetRedisDBIndex(),
			redisData.GetRedisConnectionName())
	}
}

func (t *Task) GetRedisConnectionByName(connectionName string) *redis.Pool {
	for _, redisData := range t.Object.GetRedis() {
		if redisData.GetRedisConnectionName() == connectionName {
			return redisData.GetRedisConnectionObject()
		}
	}
	panic("Please provide valid connection name")
	return nil
}

/*func (t *Task) GetRedisConnectionNames() (connectionNames []string) {
	for _,value := range t.Object.GetRedis() {
		value.GetRedisConnectionName()
	}
}*/

func (t *Task) LoadCSVData() error {
	for _, file := range t.Object.GetFiles() {
		if file.IsSerial() {
			return csv.LoadCSVFileInChannels(file.Path, file.IsEOF)
		} else {
			return csv.LoadCSVFile(file.Path)
		}
	}
	return nil
}

func (t *Task) GetFilesFromATask() []string {
	var files []string
	for _, file := range t.Object.GetFiles() {
		files = append(files, file.Name)
	}
	return files
}

func (t *Task) GetConcurrency() int {
	if t.Object.Concurrency > 0 {
		return t.Object.Concurrency
	}
	return t.Object.Rate
}

func (t *Task) GetRampUpTime() int {
	if t.Object.RampUpTime > 0 {
		return t.GetConcurrency() / t.Object.RampUpTime
	}
	return 0
}

func (t *Task) GetThinkTime() (int, int, bool) {
	if len(t.Object.ThinkTime) > 0 && isRange(t.Object.ThinkTime) {
		timer := t.Object.ThinkTime
		timer = strings.ReplaceAll(timer, constants.TIMER_RANGE, ``)
		timer = strings.ReplaceAll(timer, `)`, ``)
		values := strings.Split(timer, constants.TIMER_SPLIT)

		if len(values) <= 1 {
			panic(errors.New(constants.INPUT_TIMER_ERROR))
		}
		min, _ := strconv.Atoi(values[0])
		max, _ := strconv.Atoi(values[1])
		return min, max, true
	} else if len(t.Object.ThinkTime) > 0 {
		timer, err := strconv.Atoi(t.Object.ThinkTime)
		if err != nil {
			panic(errors.New(constants.INPUT_TIMER_ERROR))
		}
		return timer, 0, false
	}
	return 0, 0, false
}

func isRange(value string) bool {
	if strings.Contains(value, constants.TIMER_RANGE) {
		return true
	}
	return false
}

func (t *Task) GetValueFromFile(fileName string) ([]string, error) {
	for _, value := range t.Object.GetFiles() {
		if value.IsSerialRead && value.Name == fileName {
			return csv.GetValueFromFileSerially(value.GetFilePath(), value.GetValueLineSeperator())
		} else if value.Name == fileName {
			return csv.GetRandomValueFromFile(value.GetFilePath(), value.GetValueLineSeperator())
		}
	}

	return nil, errors.New("csv with name: " + fileName + "not found")
}

func (http *Httpendpoint) GetHeaderMap() map[string]string {

	var headerMap map[string]string
	var delimiter = ":"

	for _, value := range http.Headers {
		headers := strings.Split(value, delimiter)
		headerMap[headers[0]] = headers[1]
	}

	return headerMap

}

func (http *Httpendpoint) GetHeaders() []string {
	return http.Headers
}

type ResourceConfig struct {
	Genericduration int    `json:"genericduration"`
	Statsdhost      string `json:"statsdhost"`
	Stasdport       int    `json:"stasdport"`
	Filesystem      struct {
		Filename string `json:"filename"`
		Filepath string `json:"filepath"`
	} `json:"filesystem"`
	Redis struct {
		Db        string   `json:"db"`
		Instances []string `json:"instances"`
		RedisPath string   `json:"redisPath"`
	} `json:"redis"`
}

type EndPoint struct {
	BODY        string          `json:"body,omitempty"`
	URL         string          `json:"url,omitempty"`
	METHOD      string          `json:"method,omitempty"`
	PAYLOADPATH string          `json:"payloadPath,omitempty"`
	Header      EndPointHeaders `json:"headers,omitempty"`
	Cookie      EndPointCookies `json:"cookies,omitempty"`
	Response    []Response      `json:"response,omitempty"`
}

type EndPointHeaders struct {
	UserAgent      string `json:"UserAgent,omitempty"`
	ContentType    string `json:"content-type,omitempty"`
	AcceptEncoding string `json:"Accept-Encoding,omitempty"`
	Accept         string `json:"Accept,omitempty"`
}

type EndPointCookies struct {
	COOKIEONE string `json:"cookie,omitempty"`
}

type Response struct {
	JsonPath string `json:"jsonpath,omitempty"`
	Variable string `json:"variable,omitempty"`
	Index    string `json:"index,omitempty"`
	Path     string `json:"path,omitempty"`
	Key      string `json:"key,omitempty"`
}

var taskObj map[string]TaskObject

func GetTaskObjMap() map[string]TaskObject {
	if taskObj == nil {
		taskObj = make(map[string]TaskObject)
	}
	return taskObj
}

//Functions for Task receiver
func (t *Task) IsEnabled() bool {
	if t.Object.Rate == 0 {
		return false
	}
	return true
}

//func (t *Task) HasDependsOn() bool {
//	if t.Object.Dependson.Taskname == "" {
//		return false
//	}
//	return true
//}
//
//func (t *Task) HasFollowOn() bool {
//	if t.Object.Followon.Taskname == "" {
//		return false
//	}
//	return true
//}

func (t *TaskObject) GetFiles() []File {
	return t.File
}

func (file *File) GetFileName() string {
	return file.Name
}

func (file *File) GetFilePath() string {
	return file.Path
}

func (file *File) GetFileLineSeperator() string {
	return file.LineSeperator
}

func (file *File) GetValueLineSeperator() string {
	return file.ValueSeperator
}

func (file *File) IsSerial() bool {
	return file.IsSerialRead
}

func (t *TaskObject) GetRedis() []Redis {
	return t.Redis
}

func (d *Redis) GetRedisConnectionUrl() string {
	return d.Host
}

func (d *Redis) GetRedisDBIndex() int {
	return d.Index
}

func (d *Redis) GetRedisConnectionName() string {
	return d.Name
}

func (d *Redis) GetRedisConnectionObject() *redis.Pool {
	return d.Connection
}

func (t *Task) GetIndividualTaskDuration() int {
	return t.Object.Duration
}

//get global duration from resources.yml
func (r *ResourceConfig) GetGlobalDuration() int {
	return r.Genericduration
}

//get StassD Host
func (r *ResourceConfig) GetStatsDUrl() string {
	return r.Statsdhost
}

//get StassD Host
func (r *ResourceConfig) GetStatsDPost() int {
	return r.Stasdport
}

//FileSystem Map
func (r *ResourceConfig) GetFileSystemMap() map[string]string {
	var fileSysMap map[string]string
	fileSysMap["filename"] = r.Filesystem.Filename
	fileSysMap["filepath"] = r.Filesystem.Filepath
	return fileSysMap
}

//Redis Map
func (r *ResourceConfig) GetRedisMap() map[string]string {
	var redisMap map[string]string
	redisMap["redispath"] = r.Redis.RedisPath
	redisMap["redisdb"] = r.Redis.Db
	redisMap["redisinstances"] = strings.Join(r.Redis.Instances, ",")
	return redisMap
}

//func (t *Task) GetFollowOnChan() chan string {
//	return t.Object.Followon.Channel
//}
//
//func (t *Task) GetDependsOnChan() chan string {
//	return t.Object.Dependson.Channel
//}
//
//func (t *Task) GetInputChan() chan string {
//	genChan := make(chan string, 20000)
//	t.SetDependsOnChan(genChan)
//	return genChan
//}
//
//func (t *Task) GetNewChan() chan string {
//	genChan := make(chan string, 20000)
//	return genChan
//}
//
//func (t *Task) GetOutputChan() chan string {
//	var genChan chan string
//	if t.HasDependsOn() {
//		var task = GetTaskObjBasedonName(t.Object.Dependson.Taskname)
//		genChan = task.GetFollowOnChan()
//		t.SetDependsOnChan(genChan)
//	} else {
//		genChan = make(chan string, 20000)
//	}
//	return genChan
//}

// func (t *Task) GetChanForTaskGen() chan interface{} {
//      if(t.HasDependsOn()) {
// 		 return
// 	 }
// }

// func (t *Task) SetFollowOnChan() chan interface{} {
// 	followOnChannel := make(chan interface{})
// 	if t.Followon.Channelname == "" {
// 		return nil
// 	}
// 	return followOnChannel
// }

//func (t *Task) SetDependsOnChan(channel chan string) {
//	if channel != nil {
//		t.Object.Dependson.Channel = channel
//	} else {
//		genChannel := make(chan string, 20000)
//		t.Object.Dependson.Channel = genChannel
//	}
//}
//
//func (t *Task) SetFollowOnChan(channel chan string) {
//	if channel != nil {
//		t.Object.Dependson.Channel = channel
//	} else {
//		genChannel := make(chan string, 20000)
//		t.Object.Followon.Channel = genChannel
//	}
//}

type CSVObj struct {
	Data interface{}
}

type SignUpResponse struct {
	StatusCode    int           `json:"statusCode"`
	StatusMessage string        `json:"statusMessage"`
	Data          []interface{} `json:"data"`
	Tid           string        `json:"tid"`
	Sid           string        `json:"sid"`
	DeviceID      string        `json:"deviceId"`
}
