package actions

import (
	"errors"
	"fmt"
	"goranger/config"
	parsers "goranger/configparsers"
	"goranger/constants"
	"goranger/genericUtilities"
	"goranger/request"
	"log"
	"math/rand"
	"regexp"
	"strconv"
	"strings"
	"time"
)

func SampleRequest(taskObj parsers.Task) {
	fmt.Println(time.Now())
	source := rand.NewSource(time.Now().UnixNano())
	time.Sleep(time.Duration(rand.New(source).Intn(100)+200) * time.Millisecond)

}

var BaseFunc = func(taskObj parsers.Task) {
	fileValueMap := getFileValueMapForATask(taskObj, taskObj.GetFilesFromATask())
	headers, payload, url, replaceErr := ReplacePlaceHolders(taskObj, fileValueMap)
	if replaceErr != nil {
		log.Println(taskObj.Name + " replace Error " + replaceErr.Error())
	} else {
		respByte, _, err := request.MakeHTTPRequest(url, taskObj.Object.Httpendpoint.Method, headers, payload)
		printRequestDetails(payload, headers, respByte)
		if err != nil {
			go taskObj.Object.TaskCount.IncrementError()
			genericUtilities.Increment(taskObj.Name + constants.REQUEST_ERROR_COUNT)
		} else {
			go taskObj.Object.TaskCount.IncrementSuccess()
		}
	}

}

func ReplacePlaceHolders(taskObj parsers.Task, fileValueMap map[string][]string) (headers []string, payload string, url string, err error) {
	errorText := ""
	regex := regexp.MustCompile(`\${.*?}`)
	url, urlErr := replaceString(regex, fileValueMap, taskObj.Object.Httpendpoint.URL)
	if urlErr != nil {
		errorText = urlErr.Error() + "--"
	}
	payload, payloadErr := replaceString(regex, fileValueMap, genericUtilities.GetPayload(taskObj.Object.Httpendpoint.PayloadPath))
	if payloadErr != nil {
		errorText = errorText + payloadErr.Error() + "--"
	}
	headers, headersErr := replaceHeaders(taskObj, regex, fileValueMap)
	if headersErr != nil {
		errorText = errorText + headersErr.Error() + "--"
	}
	if errorText != "" && len(errorText) > 0 {
		err = errors.New(errorText)
	}
	return
}

func printRequestDetails(payload string, headers []string, respByte []byte) {
	if config.IsDebug() {
		fmt.Println("Headers --->" + strings.Join(headers, ","))
		fmt.Println("Payload --->" + payload)
		fmt.Println("Response --->" + string(respByte))
	}
}

func getFileValueMapForATask(taskObj parsers.Task, fileNames []string) map[string][]string {

	fileValueMap := make(map[string][]string)
	for _, file := range fileNames {
		value, _ := taskObj.GetValueFromFile(file)
		fileValueMap[file] = value
	}

	return fileValueMap
}
func replaceHeaders(taskObj parsers.Task, regex *regexp.Regexp, fileValueMap map[string][]string) (replacedHeaders []string, err error) {
	headers := taskObj.Object.Httpendpoint.GetHeaders()
	for _, header := range headers {
		if strings.Contains(header, `${`) {
			str, err1 := replaceString(regex, fileValueMap, header)
			if err1 != nil {
				err = err1
				return
			}
			replacedHeaders = append(replacedHeaders, str)
		} else {
			replacedHeaders = append(replacedHeaders, header)
		}
	}
	return
}

func replaceString(regex *regexp.Regexp, fileValueMap map[string][]string, str string) (string, error) {
	for _, placeHolder := range getPlaceHolders(regex, str) {
		fileName := getFileName(placeHolder)
		index := getIndex(placeHolder, fileName)
		value := fileValueMap[fileName]
		if index > len(value) {
			return "", errors.New("Index not present: trying to access " + strconv.Itoa(index) + "th index in file: " + fileName + " for the replacement of placeholder: " + placeHolder)
		}
		str = strings.ReplaceAll(str, placeHolder, value[index])
	}
	return str, nil
}

func getFileName(str string) string {
	end := strings.Index(str, `_`)
	return str[2:end]
}

func getIndex(str string, filename string) int {
	start := strings.Index(str, `_`)
	end := strings.Index(str, `}`)
	index, _ := strconv.Atoi(str[start+1 : end])
	return index
}

func getPlaceHolders(regex *regexp.Regexp, str string) []string {
	return regex.FindAllString(str, -1)
}
