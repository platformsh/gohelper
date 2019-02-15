package gohelper

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"testing"
)

type envList map[string]string

func TestNotOnPlatformReturnsCorrectly(t *testing.T) {

	_, err := NewConfigReal(nonPlatformEnv())

	if err == nil {
		t.Fail()
	}
}

func TestInBuildReturnsTrueInBuild(t *testing.T) {

	config, err := NewConfigReal(buildEnv(envList{}))
	ok(t, err)

	if !config.InBuild() {
		t.Fail()
	}
}

func TestInBuildReturnsFalseInRumtime(t *testing.T) {

	config, err := NewConfigReal(runtimeEnv(envList{}))
	ok(t, err)

	if config.InBuild() {
		t.Fail()
	}
}

func TestOnEnterpriseReturnsTrueOnEnterprise(t *testing.T) {
	config, err := NewConfigReal(runtimeEnv(envList{"PLATFORM_MODE": "enterprise"}))
	ok(t, err)

	if !config.OnEnterprise() {
		t.Fail()
	}
}

func TestOnEnterpriseReturnsFalseOnStandard(t *testing.T) {
	config, err := NewConfigReal(runtimeEnv(envList{}))
	ok(t, err)

	if config.OnEnterprise() {
		t.Fail()
	}
}

// This function produces a getter of the same signature as os.Getenv() that
// always returns an empty string, simulating a non-Platform environment.
func nonPlatformEnv() func(string) string {
	return func(key string) string {
		return ""
	}
}

// This function produces a getter of the same signature as os.gGetenv()
// that returns test values to simulate a build environment.
func buildEnv(env envList) func(string) string {

	// Create build time env.
	vars := loadJsonFile("testdata/ENV.json")
	env = mergeMaps(env, vars)
	env["PLATFORM_VARIABLES"] = encodeJsonFile("testdata/PLATFORM_VARIABLES.json")
	env["PLATFORM_APPLICATION"] = encodeJsonFile("testdata/PLATFORM_APPLICATION.json")

	return func(key string) string {
		if val, ok := env[key]; ok {
			return val
		} else {
			return ""
		}
	}
}

// This function produces a getter of the same signature as os.gGetenv()
// that returns test values to simulate a runtime environment.
func runtimeEnv(env envList) func(string) string {

	// Create runtimeVars env.
	vars := loadJsonFile("testdata/ENV.json")
	env = mergeMaps(env, vars)
	env["PLATFORM_VARIABLES"] = encodeJsonFile("testdata/PLATFORM_VARIABLES.json")
	env["PLATFORM_APPLICATION"] = encodeJsonFile("testdata/PLATFORM_APPLICATION.json")
	env["PLATFORM_RELATIONSHIPS"] = encodeJsonFile("testdata/PLATFORM_RELATIONSHIPS.json")
	env["PLATFORM_ROUTES"] = encodeJsonFile("testdata/PLATFORM_ROUTES.json")

	runtimeVars := loadJsonFile("testdata/ENV_runtime.json")
	env = mergeMaps(env, runtimeVars)

	return func(key string) string {
		if val, ok := env[key]; ok {
			return val
		} else {
			return ""
		}
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
