package ntnx_api_call

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Emoji symbols from http://www.unicode.org/emoji/charts/emoji-list.html
var symbols = map[string]string{
	"FAIL":    "\U0000274C",
	"INFO":    "\U0001F449",
	"OK":      "\U00002705",
	"WAIT":    "\U0001F55B",
	"NEUTRAL": "\U00002796",
}

type Ntnx_endpoint struct {
	PC            string
	PE            string
	Mode          string
	User          string
	Password      string
	Cert          string
	Chain         string
	Key           string
	DebugMode     bool
	DebugLogger   *zap.SugaredLogger
	RetryMode     bool
	RetryNumber   int
	Compatibility bool
}

const const_max_retry int = 3

// =========== CheckErr ===========
// This function is will handle errors
func CheckErr(context string, err error, debugmode bool, debuglogger *zap.SugaredLogger) {
	if err != nil {
		if debugmode {
			debuglogger.Debugf("API Call Error : %s\n", err.Error())
		}
		log.Fatal(symbols["FAIL"], context, err.Error())
	}
}

// =========== WaitForTask ===========
// Wait for end of task and return stats

func (e Ntnx_endpoint) WaitForTask(task string) (bool, string, string) {
	url := "/api/nutanix/v3/tasks/" + task

	type TmpStruct struct {
		ProgressMessage    string `json:"progress_message"`
		PercentageComplete int64  `json:"percentage_complete"`
		Status             string `json:"status"`
		ErrorCode          string `json:"error_code"`
		ErrorDetail        string `json:"error_detail"`
	}

	var ReturnValue TmpStruct

	for ReturnValue.PercentageComplete < 100 {
		e.CallAPIJSON("PC", "GET", url, "", &ReturnValue)
		time.Sleep(time.Duration(10) * time.Second)
	}

	if ReturnValue.Status == "SUCCEEDED" {
		return true, ReturnValue.Status, ReturnValue.ErrorDetail
	} else {
		return false, ReturnValue.Status, ReturnValue.ErrorDetail
	}

}

// =========== ActivateDebug ===========
// Create file to store all API calls for debug purpose
func (e *Ntnx_endpoint) ActivateDebug(filename string) {
	myconf := zap.NewDevelopmentEncoderConfig()
	myconf.TimeKey = "timestamp"
	myconf.EncodeTime = zapcore.ISO8601TimeEncoder
	myconf.CallerKey = ""

	config := zap.Config{
		Level:            zap.NewAtomicLevelAt(zapcore.DebugLevel),
		Development:      true,
		Encoding:         "console",
		EncoderConfig:    myconf,
		OutputPaths:      []string{filename},
		ErrorOutputPaths: []string{filename},
	}

	logger, err := config.Build()
	if err != nil {
		panic(err)
	}
	e.DebugMode = true
	e.DebugLogger = logger.Sugar()
	defer logger.Sync()
}

// =========== CallAPIJSON ===========
// Do a call API and unmarshall the result
func (e Ntnx_endpoint) CallAPIJSON(target string, method string, url string, payload string, retValue interface{}) {

	var long_url, ReqMethod string
	var jsonStr []byte
	var resp *http.Response
	client := &http.Client{}

	if strings.ToUpper(target) == "PE" {
		long_url = "https://" + e.PE + ":9440" + url
	} else {
		long_url = "https://" + e.PC + ":9440" + url
	}

	if strings.ToUpper(method) == "POST" {
		jsonStr = []byte(payload)
		ReqMethod = http.MethodPost
	} else if strings.ToUpper(method) == "GET" {
		jsonStr = nil
		ReqMethod = http.MethodGet
	} else if strings.ToUpper(method) == "PATCH" {
		jsonStr = []byte(payload)
		ReqMethod = http.MethodPatch
	} else if strings.ToUpper(method) == "DELETE" {
		jsonStr = []byte(payload)
		ReqMethod = http.MethodDelete
	} else {
		log.Fatalln("HTTP method", method, "not handled")
	}

	// Create new request

	// Log it
	if e.DebugMode {
		e.DebugLogger.Debugf("New API Call : %s %s %s\n", ReqMethod, long_url, bytes.NewBuffer(jsonStr))
	}

	// Execute it
	req, err := http.NewRequest(ReqMethod, long_url, bytes.NewBuffer(jsonStr))
	CheckErr("Unable to prepare API Call", err, e.DebugMode, e.DebugLogger)

	// Define default headers
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")

	if e.Mode == "password" {
		// Authentication
		req.SetBasicAuth(e.User, e.Password)

	} else if e.Mode == "cert" {

		_, err := tls.X509KeyPair([]byte(e.Cert), []byte(e.Key))
		CheckErr("Unable to load certs", err, e.DebugMode, e.DebugLogger)

	} else {
		log.Fatalln("FAIL", "Mode "+e.Mode+" unknown for Nutanix API call")
	}

	// Launch request
	resp, err = client.Do(req)

	// If an error occurs, we retry
	for retry := 1; retry < const_max_retry && err != nil; retry++ {
		if e.DebugMode {
			e.DebugLogger.Debugf("Try %d...\n", retry)
		}
		resp, err = client.Do(req)
	}
	CheckErr("API Call failed", err, e.DebugMode, e.DebugLogger)

	// Log error is > 299
	if int(resp.StatusCode) > 299 {

		// Put it in debug file if requested
		if e.DebugMode {
			e.DebugLogger.Debugf("API Call Error : %s\n%s\n", resp.Status, resp.Request)
		}

		// Display it and exit
		log.Fatal(symbols["FAIL"], " API Call failed :", resp.Status)
	}

	defer resp.Body.Close()
	bodyBytes, err := io.ReadAll(resp.Body)
	CheckErr("Unable to read API answer body", err, e.DebugMode, e.DebugLogger)

	// Transform json answer to map
	err = json.Unmarshal(bodyBytes, &retValue)
	CheckErr("Unable to get json answer from API Call.", err, e.DebugMode, e.DebugLogger)
}
