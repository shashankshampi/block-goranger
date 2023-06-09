package csv

import (
	"errors"
	"goranger/genericUtilities"
	"math/rand"
	"strings"
	"time"
)

const (
	EOF_ERROR = "Reached EOF"
)

var CSVMap = make(map[string][]string)
var CSVChannelMap = make(map[string]chan string)

func LoadCSVFile(csvPath string) error {

	_, isKeyPresent := CSVMap[csvPath]
	if !isKeyPresent {
		values, message, err := genericUtilities.ParseFile(csvPath)
		if err != nil {
			return errors.New(message)
		}
		CSVMap[csvPath] = values
	}

	return nil
}

func GetRandomValueFromFile(csvPath string, valueSeperator string) ([]string, error) {

	local := rand.New(rand.NewSource(time.Now().UnixNano()))
	values := CSVMap[csvPath]
	value := values[local.Intn(len(values))]
	finalValues := strings.Split(value, valueSeperator)
	if len(finalValues) == 0 {
		return finalValues, errors.New("no value read from csv file")
	}
	return finalValues, nil
}

func LoadCSVFileInChannels(csvPath string, isEOF bool) error {

	_, isKeyPresent := CSVChannelMap[csvPath]
	if !isKeyPresent {
		values, message, err := genericUtilities.ParseFile(csvPath)
		if err != nil {
			return errors.New(message)
		}
		channel := make(chan string, len(values))
		CSVChannelMap[csvPath] = channel

		go func() {
			for i := 0; i < len(values); i++ {
				channel <- values[i]

				if i == len(values)-1 && !isEOF {
					i = 0 - 1
				}
			}
		}()
	}
	return nil
}

func GetValueFromFileSerially(csvPath string, valueSeperator string) ([]string, error) {

	for value := range CSVChannelMap[csvPath] {
		return strings.Split(value, valueSeperator), nil
	}
	return nil, errors.New(EOF_ERROR)
}
