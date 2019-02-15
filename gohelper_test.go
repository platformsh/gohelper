package gohelper

import (
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"testing"
)

type envList map[string]string

var mockEnvironmentBuild envList

var mockEnvironmentRuntime envList

func TestMain(m *testing.M) {

	// Create build time env.
	mockEnvironmentBuild = loadJsonFile("testdata/ENV.json")
	mockEnvironmentBuild["PLATFORM_VARIABLES"] = encodeJsonFile("testdata/PLATFORM_VARIABLES.json")
	mockEnvironmentBuild["PLATFORM_APPLICATION"] = encodeJsonFile("testdata/PLATFORM_APPLICATION.json")

	// Create runtimeVars env.
	mockEnvironmentRuntime = loadJsonFile("testdata/ENV.json")
	mockEnvironmentRuntime["PLATFORM_VARIABLES"] = encodeJsonFile("testdata/PLATFORM_VARIABLES.json")
	mockEnvironmentRuntime["PLATFORM_APPLICATION"] = encodeJsonFile("testdata/PLATFORM_APPLICATION.json")
	mockEnvironmentRuntime["PLATFORM_RELATIONSHIPS"] = encodeJsonFile("testdata/PLATFORM_RELATIONSHIPS.json")
	mockEnvironmentRuntime["PLATFORM_ROUTES"] = encodeJsonFile("testdata/PLATFORM_ROUTES.json")

	runtimeVars := loadJsonFile("testdata/ENV_runtime.json")
	mockEnvironmentRuntime = mergeMaps(mockEnvironmentRuntime, runtimeVars)

	//spew.Dump(getKeys(mockEnvironmentBuild))
	//fmt.Println("-----------------------------")
	//spew.Dump(getKeys(mockEnvironmentRuntime))

	flag.Parse()
	exitCode := m.Run()

	os.Exit(exitCode)
}

func TestNotOnPlatformReturnsCorrectly(t *testing.T) {
	os.Clearenv()

	_, err := NewConfig()

	if err == nil {
		t.Fail()
	}
}

func populateBuildEnvironment() func() {
	os.Clearenv()

	for k, v := range mockEnvironmentBuild {
		os.Setenv(k, v)
	}

	return func() {
		os.Clearenv()
	}
}

func getKeys(data envList) []string {
	keys := make([]string, 0)
	for key := range data {
		keys = append(keys, key)
	}

	return keys
}

func mergeMaps(a envList, b envList) envList {
	for k, v := range b {
		a[k] = v
	}
	return a
}

func encodeJsonFile(file string) string {
	jsonFile, err := os.Open(file)

	if err != nil {
		fmt.Println(err)
	}
	defer jsonFile.Close()

	byteValue, _ := ioutil.ReadAll(jsonFile)

	val := base64.StdEncoding.EncodeToString(byteValue)
	return val
}

func loadJsonFile(file string) envList {
	jsonFile, err := os.Open(file)

	if err != nil {
		fmt.Println(err)
	}
	defer jsonFile.Close()

	byteValue, _ := ioutil.ReadAll(jsonFile)

	var result envList
	json.Unmarshal([]byte(byteValue), &result)

	return result
}

// These utilities copied with permission from:
// https://github.com/benbjohnson/testing

// assert fails the test if the condition is false.
func assert(tb testing.TB, condition bool, msg string, v ...interface{}) {
	if !condition {
		_, file, line, _ := runtime.Caller(1)
		fmt.Printf("\033[31m%s:%d: "+msg+"\033[39m\n\n", append([]interface{}{filepath.Base(file), line}, v...)...)
		tb.FailNow()
	}
}

// ok fails the test if an err is not nil.
func ok(tb testing.TB, err error) {
	if err != nil {
		_, file, line, _ := runtime.Caller(1)
		fmt.Printf("\033[31m%s:%d: unexpected error: %s\033[39m\n\n", filepath.Base(file), line, err.Error())
		tb.FailNow()
	}
}

// equals fails the test if exp is not equal to act.
func equals(tb testing.TB, exp, act interface{}) {
	if !reflect.DeepEqual(exp, act) {
		_, file, line, _ := runtime.Caller(1)
		fmt.Printf("\033[31m%s:%d:\n\n\texp: %#v\n\n\tgot: %#v\033[39m\n\n", filepath.Base(file), line, exp, act)
		tb.FailNow()
	}
}
