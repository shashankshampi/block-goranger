package helpers

import (
	"errors"
	"github.com/tidwall/gjson"
	"io/ioutil"
)

type JSON struct {
	json string
}

func LoadJSON(filePath string) (*JSON,error) {
	file, err := ioutil.ReadFile(filePath)

	if err != nil {
		return nil, errors.New("Not able to read the json file on " + filePath)
	}
	return &JSON {
		json : string(file),
	}, nil
}

func (json *JSON) Get(pattern string) gjson.Result {
	return gjson.Get(json.json, pattern)
}

