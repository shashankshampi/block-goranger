package main

import (
	"fmt"
	"github.com/fsnotify/fsnotify"
	"goranger/actions"
	"goranger/genericUtilities"
	"goranger/lib"
	"log"
	"os"

	//"github.com/fsnotify/fsnotify"
	"github.com/gin-gonic/gin"
	"github.com/hpcloud/tail"
	"io/ioutil"
	"net/http"
	"os/exec"
	"strconv"
	"strings"
)

var session *Session

type Session struct {
	hudor       *lib.Hudor
	branch      []string
	profile     []string
	taskList    []string
	host        string
	pod         string
	fileChannel chan string
	logOutput   []string
	logOffset   int64
}

const (
	results = `{
"title" : "${title}",
"success" : "${success}",
"rate" : "${rate}",
"error" : "${error}"
}`
)

var TailConf = tail.Config{
	Follow: true,
}

var TailObj *tail.Tail

type Result struct {
	title   string `json:"title"`
	success string `json:"success"`
	rate    string `json:"rate"`
	error   string `json:"error"`
}

func resetSession() {
	os.Remove("stdout.log")
	file, _ := os.Create("stdout.log")
	log.SetOutput(file)
	session = &Session{
		branch:      []string{},
		profile:     []string{},
		taskList:    []string{},
		hudor:       nil,
		host:        "",
		pod:         "",
		fileChannel: make(chan string, 100000),
		logOutput:   []string{},
		logOffset:   0,
	}
}

func main() {
	router := gin.Default()
	resetSession()
	file, _ := os.Create("stdout.log")
	log.SetOutput(file)
	v1 := router.Group("/api/v1/hudor")
	{
		v1.GET("/getticket", getTicket)
		v1.GET("/gethost", getHost)
		v1.GET("/getpod", getTicket)
		v1.GET("/getsession", getSession)
		v1.GET("/loadprofile/:name", loadProfile)
		v1.GET("/getprofiles", getProfiles)
		v1.GET("/getbranches", getBranches)
		v1.GET("/checkoutbranch/:name", checkoutBranch)
		v1.GET("/getalltasks/:profile", getAllTasks)
		v1.GET("/runtask/:tasks", runTasks)
		v1.GET("/getActiveTasks", getActiveTasks)
		v1.GET("/updateTaskRate/:task/:rate", updateTaskRate)
		v1.GET("/stopTask/:tasks", stopTask)
		v1.GET("/getresults", getResults)
		v1.GET("/reset", reset)
		v1.GET("/logs", logs)
		v1.POST("/upload/payload", uploadPayload)
		v1.POST("/upload/csv", uploadCsv)
		v1.POST("/upload/task", uploadTask)

	}
	router.Run(":8080")
}

func setResponseHeaders(c *gin.Context) *gin.Context {
	c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
	c.Header("Access-Control-Allow-Origin", "*")
	c.Request.Header.Set("Access-Control-Allow-Origin", "*")

	return c
}

func getTicket(c *gin.Context) {

	var values = []string{}

	values = append(values, "PER-377")
	values = append(values, "PER-354")

	c.Request.Header.Set("Access-Control-Allow-Origin", "*")
	c = setResponseHeaders(c)
	c.JSON(http.StatusOK, values)
}

func getHost(c *gin.Context) {

	var values = []string{}

	values = append(values, "18.212.188.226")

	c.Request.Header.Set("Access-Control-Allow-Origin", "*")
	c = setResponseHeaders(c)
	c.JSON(http.StatusOK, values)
}

func getSession(c *gin.Context) {

	var values = [][]string{}

	var host = []string{}
	session.host = "18.212.188.226"

	if session.hudor != nil && len(session.hudor.ActiveTaskMap) > 0 {
		values = append(values, session.branch)
		values = append(values, session.profile)
		values = append(values, session.taskList)
		values = append(values, append(host, session.host))
	}
	c.Request.Header.Set("Access-Control-Allow-Origin", "*")
	c = setResponseHeaders(c)
	c.JSON(http.StatusOK, values)
}
func getProfiles(c *gin.Context) {
	files, err := ioutil.ReadDir("profiles/")
	var fileNames = []string{}
	branch := c.Query("selectedBranch")

	branches := []string{}
	session.branch = append(branches, branch)
	for _, v := range files {
		fileNames = append(fileNames, strings.ReplaceAll(v.Name(), `.json`, ``))
	}

	c = setResponseHeaders(c)
	if err == nil {
		c.JSON(http.StatusOK, fileNames)
	} else {
		log.Println(err.Error())
		c.JSON(http.StatusOK, fileNames)
	}

}

func getBranches(c *gin.Context) {
	c.Request.Header.Set("Access-Control-Allow-Origin", "*")

	var branches []string
	branches = append(branches, "vertical_perf")

	c = setResponseHeaders(c)
	c.JSON(http.StatusOK, branches)

}

func checkoutBranch(c *gin.Context) {
	//name:=c.Param("name")
	out, err := exec.Command("cd", `go env | grep GOPATH | cut -d '=' -f 2 | tr -d '"'`).Output()
	//out1, _ := exec.Command(`pwd`).Output()
	//_, err1 := exec.Command(`git pull`).Output()
	//_, err2 := exec.Command(`git checkout ` + name).Output()

	c = setResponseHeaders(c)
	if err != nil {
		c.JSON(http.StatusOK, string(out))
	} else {
		c.JSON(http.StatusInternalServerError, "error")
	}
}

func loadProfile(c *gin.Context) {
	name := c.Param("name")
	var err error
	hudorObj, err := lib.Init("profiles/" + name + ".json")
	session.hudor = hudorObj
	if err == nil {
		c.JSON(http.StatusOK, "")
	} else {
		c.JSON(http.StatusInternalServerError, "")
	}
}

func getAllTasks(c *gin.Context) {
	TailObj, _ = tail.TailFile("stdout.log", tail.Config{Follow: true})
	go func() {
		for lines := range TailObj.Lines {
			session.logOutput = append(session.logOutput, lines.Text)
		}
	}()
	profile := c.Param("profile")
	var tasks = []string{}
	var err error
	c = setResponseHeaders(c)
	profiles := []string{}
	session.profile = append(profiles, profile)
	hudorObj, err := lib.Init("profiles/" + profile + ".json")
	session.hudor = hudorObj
	session.taskList = tasks
	if err == nil {
		for name, _ := range session.hudor.TaskObjectMap {
			tasks = append(tasks, name)
		}
		c.JSON(http.StatusOK, tasks)
		session.taskList = tasks
	} else {
		c.JSON(http.StatusOK, tasks)
	}

}

func runTasks(c *gin.Context) {

	param := c.Param("tasks")
	if strings.Contains(param, ",") {
		tasks := strings.Split(param, ",")
		for _, task := range tasks {
			session.hudor.Run(task, actions.BaseFunc)
		}
	} else {
		session.hudor.Run(param, actions.BaseFunc)
	}
	c = setResponseHeaders(c)
	c.JSON(http.StatusOK, "")
}

func getActiveTasks(c *gin.Context) {
	var tasks = []string{}

	if session.hudor != nil {
		for name, _ := range session.hudor.ActiveTaskMap {
			tasks = append(tasks, name)
		}
	}
	c = setResponseHeaders(c)
	c.JSON(http.StatusOK, tasks)
}

func updateTaskRate(c *gin.Context) {
	task := c.Param("task")
	param := c.Param("rate")
	rate, _ := strconv.Atoi(param)
	err := session.hudor.UpdateTaskRate(task, rate)
	c = setResponseHeaders(c)

	if err == nil {
		c.JSON(http.StatusOK, "200")
	} else {
		c.JSON(http.StatusOK, "500")
	}
}

func stopTask(c *gin.Context) {
	param := c.Param("tasks")
	if strings.Contains(param, ",") {
		tasks := strings.Split(param, ",")
		for _, task := range tasks {
			session.hudor.StopTask(task)
		}
	} else {
		session.hudor.StopTask(param)
	}
	c = setResponseHeaders(c)
	c.JSON(http.StatusOK, "200")
}

func reset(c *gin.Context) {

	if session.hudor != nil && len(session.hudor.ActiveTaskMap) > 0 {
		session.hudor.StopAllTasks()
		session.hudor.Close()
	}
	resetSession()
	c = setResponseHeaders(c)
	c.JSON(http.StatusOK, "")
}

func getResults(c *gin.Context) {

	var tasks = []string{}
	if session.hudor != nil {
		for name, _ := range session.hudor.ActiveTaskMap {
			temp := results
			temp = strings.ReplaceAll(temp, `${title}`, name)
			temp = strings.ReplaceAll(temp, `${success}`, strconv.Itoa(session.hudor.TaskObjectMap[name].Object.TaskCount.GetSuccessCount()))
			temp = strings.ReplaceAll(temp, `${rate}`, strconv.Itoa(session.hudor.TaskObjectMap[name].Object.Rate))
			temp = strings.ReplaceAll(temp, `${error}`, strconv.Itoa(session.hudor.TaskObjectMap[name].Object.TaskCount.GetErrorCount()))
			tasks = append(tasks, temp)
		}
	}
	c = setResponseHeaders(c)
	c.JSON(http.StatusOK, tasks)
}

func logs(c *gin.Context) {

	var output = ""

	i := 0
	for i = 0; i < 10; i++ {
		//fmt.Println(len(session.logOutput))
		//fmt.Println(int64(i) + session.logOffset)
		if int64(len(session.logOutput)) > int64(i)+session.logOffset {
			//output=append(output,session.logOutput[int64(i) + session.logOffset])
			output = output + session.logOutput[int64(i)+session.logOffset] + "\n"
		} else {
			break
		}
	}
	session.logOffset = session.logOffset + int64(i)

	c = setResponseHeaders(c)
	c.JSON(http.StatusOK, output)

}

func uploadPayload(c *gin.Context) {
	c = setResponseHeaders(c)
	c.Request.Header.Add("Content-Type", "multipart/form-data")
	form, err := c.MultipartForm()
	if err != nil {
		fmt.Println(err.Error())
		c.JSON(http.StatusBadRequest, "400")
		return
	}

	files := form.File["file"]
	for _, file := range files {
		// Upload the file to specific dst.
		uploadErr := c.SaveUploadedFile(file, "payloads/"+file.Filename)
		if uploadErr != nil {
			fmt.Println(uploadErr.Error())
			c.String(http.StatusInternalServerError, "500")
			return
		}
	}
	c.JSON(http.StatusOK, "200")
}

func uploadCsv(c *gin.Context) {
	c = setResponseHeaders(c)
	c.Request.Header.Add("Content-Type", "multipart/form-data")
	form, err := c.MultipartForm()
	if err != nil {
		fmt.Println(err.Error())
		c.JSON(http.StatusBadRequest, "400")
		return
	}

	files := form.File["file"]
	for _, file := range files {
		// Upload the file to specific dst.
		uploadErr := c.SaveUploadedFile(file, "files/"+file.Filename)
		if uploadErr != nil {
			fmt.Println(uploadErr.Error())
			c.String(http.StatusInternalServerError, "500")
			return
		}
	}
	c.JSON(http.StatusOK, "200")
}

func uploadTask(c *gin.Context) {
	c = setResponseHeaders(c)
	c.Request.Header.Add("Content-Type", "multipart/form-data")
	form, err := c.MultipartForm()
	if err != nil {
		fmt.Println(err.Error())
		c.JSON(http.StatusBadRequest, "400")
		return
	}

	files := form.File["file"]
	for _, file := range files {
		// Upload the file to specific dst.
		uploadErr := c.SaveUploadedFile(file, "profiles/"+file.Filename)
		if uploadErr != nil {
			fmt.Println(uploadErr.Error())
			c.String(http.StatusInternalServerError, "500")
			return
		}
	}
	c.JSON(http.StatusOK, "200")
}

func tailLogs() {
	count := 0
	watcher, _ := fsnotify.NewWatcher()
	watcher.Add("stdout.log")
	<-watcher.Events
	for {
		lines, _, _ := genericUtilities.ParseFile("stdout.log")
		session.fileChannel <- lines[count]
		count++
	}
}
