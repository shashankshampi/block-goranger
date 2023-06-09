package genericUtilities

import (
	"bytes"
	"compress/gzip"
	"errors"
	"io/ioutil"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/valyala/fasthttp"
)

func GetRateLimiter(rate int64) <-chan time.Time{
	var output = int64(1e9 / rate)
	ratelimiter := time.NewTicker(time.Nanosecond * time.Duration(output)).C
	return ratelimiter
}

func GetRandomValue(ulimit int) int {
	rand.Seed(time.Now().UnixNano())
	return rand.Intn(ulimit)
}

func GetRandomIndex(length int) int {
	random := rand.New(rand.NewSource(time.Now().UnixNano())) // initialize local pseudorandom generator
	return random.Intn(length)
}

func GetValueFromRedisOutput(input string, redisKey string) string {
	var output string
	var s = strings.Split(input, redisKey+":")
	output = strings.TrimPrefix(s[1], " ")
	return output
}

func DecompressResponseBody(resp *fasthttp.Response) []byte {
	var responseBody []byte
	switch string(resp.Header.Peek("Content-Encoding")) {
	case "gzip":
		responseBody, _ = resp.BodyGunzip()
	default:
		responseBody = resp.Body()
	}
	return responseBody
}

func GetPayload(payloadPath string) string {
	if payloadPath != "" {
		dir, _ := os.Getwd()
		templateData, _ := ioutil.ReadFile(dir + "/payloads/" + payloadPath)
		return string(templateData)
	} else {
		return ""
	}
}

func GetGZipedData(data string) ([]byte, error) {
	var buf bytes.Buffer
	g := gzip.NewWriter(&buf)
	if _, err := g.Write([]byte(data)); err != nil {
		return nil, errors.New("Not able to convert")
	}
	if err := g.Close(); err != nil {
		return nil, errors.New("Not able to convert")
	}
	return buf.Bytes(), nil
}
