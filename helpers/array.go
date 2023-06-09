package helpers

import (
	"github.com/tidwall/gjson"
	"math/rand"
	"time"
)

func GetRandomValueFromAStringSlice(slice []string) string {

	if len(slice) > 0 {

		random := rand.New(rand.NewSource(time.Now().UnixNano())) // initialize local pseudorandom generator
		value := slice[random.Intn(len(slice))]

		return value
	}
	return ""

}

func GetRandomValueFromArray(slice []gjson.Result) string {

	if len(slice) > 0 {

		random := rand.New(rand.NewSource(time.Now().UnixNano())) // initialize local pseudorandom generator
		value := slice[random.Intn(len(slice))]

		return value.String()
	}
	return ""
}


