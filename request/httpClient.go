package request

import (
	"crypto/tls"
	"errors"
	"github.com/valyala/fasthttp"
	"goranger/constants"
	"log"
	"net"
	"strconv"
	"strings"
	"time"
)

const (
	DefaultTimeout           = 60 * time.Second
	DefaultConnections       = 50000
	DefaultConnectionTimeout = 300 * time.Second
)

var (
	FasthttpClient = &fasthttp.Client{
		MaxConnsPerHost:     DefaultConnections,
		MaxIdleConnDuration: DefaultConnectionTimeout,
		ReadTimeout:         DefaultTimeout,
		WriteTimeout:        DefaultTimeout,
		TLSConfig: &tls.Config{
			InsecureSkipVerify: true,
			ClientSessionCache: tls.NewLRUClientSessionCache(0),
		},
		Dial: func(addr string) (net.Conn, error) {
			var dialer = net.Dialer{
				Timeout:   60 * time.Second,
				KeepAlive: 5 * time.Second,
			}
			return dialer.Dial("tcp", addr)
		},
	}
	errMissingLocation = errors.New("missing Location header for http redirect")
)

type Httpendpoint struct {
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

func GetHttpClientObj() *fasthttp.Client {
	if FasthttpClient == nil {
		return FasthttpClient
	}
	return FasthttpClient
}

func MakeHTTPRequest(url string, methodName string, headers []string,
	payload string) ([]byte, *fasthttp.Response, error) {

	client := GetHttpClientObj()
	reqValidate := fasthttp.AcquireRequest()
	respValidate := fasthttp.AcquireResponse()

	defer fasthttp.ReleaseResponse(respValidate)
	defer fasthttp.ReleaseRequest(reqValidate)

	reqValidate.SetRequestURI(url)
	reqValidate.Header.SetMethod(methodName)

	if payload != "" {
		reqValidate.SetBody([]byte(payload))
	}

	for _, value := range headers {
		headerValue := strings.Split(value, ":")
		reqValidate.Header.Add(headerValue[0], headerValue[1])
	}

	reqValidate.Header.SetContentType("application/json")

	errValidate := client.Do(reqValidate, respValidate)

	if errValidate == nil && respValidate.StatusCode() == 200 {
		log.Println("Request: " + reqValidate.String())
		log.Println("Response: " + reqValidate.String())

		log.Println(url + "----->:SUCCESS")
	} else if errValidate == nil && respValidate.StatusCode() != 200 {
		log.Println(url+"----->:FAILED ", respValidate.StatusCode())
		errValidate = errors.New(constants.RESPONSE_CODE_NOT_200 + strconv.Itoa(respValidate.StatusCode()))
	} else if errValidate != nil {
		log.Println(url + "----->:FAILED")
		log.Println(errValidate.Error())
	}

	return respValidate.Body(), respValidate, errValidate
}
