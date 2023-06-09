package genericUtilities

import (

	"github.com/tidwall/gjson"
	"strings"
)

func ReplaceInStringArray(stringArray []string, replacePattern string, replaceString string) ([]string) {

	var array []string
	for _,value := range stringArray {
		array = append(array,strings.ReplaceAll(value, replacePattern, replaceString))
	}

	return array
}

func ReplaceTidSidTokenInHeadersFromJson(headers []string, json string) ([]string){

	values := gjson.GetMany(json,"tid","sid","token")

	headers = ReplaceInStringArray(headers,"${tid}", values[0].String())
	headers = ReplaceInStringArray(headers, "${sid}", values[1].String())
	headers = ReplaceInStringArray(headers, "${token}", values[2].String())

	return headers;
}

