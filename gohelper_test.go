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

func TestNotOnPlatformReturnsCorrectly(t *testing.T) {

	_, err := NewConfigReal(nonPlatformEnv(), "PLATFORM_")

	if err == nil {
		t.Fail()
	}
}

func TestInBuildReturnsTrueInBuild(t *testing.T) {

	config, err := NewConfigReal(buildEnv(envList{}), "PLATFORM_")
	ok(t, err)

	if !config.InBuild() {
		t.Fail()
	}
}

func TestInBuildReturnsFalseInRumtime(t *testing.T) {

	config, err := NewConfigReal(runtimeEnv(envList{}), "PLATFORM_")
	ok(t, err)

	if config.InBuild() {
		t.Fail()
	}
}

func TestInRuntimeReturnsTrueInRuntime(t *testing.T) {

	config, err := NewConfigReal(runtimeEnv(envList{}), "PLATFORM_")
	ok(t, err)

	if !config.InRuntime() {
		t.Fail()
	}
}

func TestInRuntimeReturnsFalseInBuild(t *testing.T) {

	config, err := NewConfigReal(buildEnv(envList{}), "PLATFORM_")
	ok(t, err)

	if config.InRuntime() {
		t.Fail()
	}
}

func TestOnEnterpriseReturnsTrueOnEnterprise(t *testing.T) {
	config, err := NewConfigReal(runtimeEnv(envList{"PLATFORM_MODE": "enterprise"}), "PLATFORM_")
	ok(t, err)

	if !config.OnEnterprise() {
		t.Fail()
	}
}

func TestOnEnterpriseReturnsFalseOnStandard(t *testing.T) {
	config, err := NewConfigReal(runtimeEnv(envList{}), "PLATFORM_")
	ok(t, err)

	if config.OnEnterprise() {
		t.Fail()
	}
}

func TestOnProductionOnEnterpriseProdReturnsTrue(t *testing.T) {
	config, err := NewConfigReal(runtimeEnv(envList{
		"PLATFORM_MODE":   "enterprise",
		"PLATFORM_BRANCH": "production",
	}), "PLATFORM_")
	ok(t, err)

	assert(t, config.OnProduction(), "OnProduction() returned false when it should be true.")
}

func TestOnProductionOnEnterpriseStagingReturnsFalse(t *testing.T) {
	config, err := NewConfigReal(runtimeEnv(envList{
		"PLATFORM_MODE":   "enterprise",
		"PLATFORM_BRANCH": "staging",
	}), "PLATFORM_")
	ok(t, err)

	assert(t, !config.OnProduction(), "OnProduction() returned true when it should be false.")
}

func TestOnProductionOnStandardProdReturnsTrue(t *testing.T) {
	config, err := NewConfigReal(runtimeEnv(envList{
		"PLATFORM_BRANCH": "master",
	}), "PLATFORM_")
	ok(t, err)

	assert(t, config.OnProduction(), "OnProduction() returned false when it should be true.")
}

func TestOnProductionOnStandardStagingReturnsFalse(t *testing.T) {
	config, err := NewConfigReal(runtimeEnv(envList{}), "PLATFORM_")
	ok(t, err)

	assert(t, !config.OnProduction(), "OnProduction() returned true when it should be false.")
}

func TestBuildPropertyInBuildExists(t *testing.T) {
	config, err := NewConfigReal(buildEnv(envList{}), "PLATFORM_")
	ok(t, err)

	equals(t, "/app", config.AppDir())
	equals(t, "app", config.ApplicationName())
	equals(t, "test-project", config.Project())
	equals(t, "abc123", config.TreeId())
	equals(t, "def789", config.ProjectEntropy())
}

func TestBuildAndRuntimePropertyInRuntimeExists(t *testing.T) {
	config, err := NewConfigReal(runtimeEnv(envList{}), "PLATFORM_")
	ok(t, err)

	equals(t, "/app", config.AppDir())
	equals(t, "app", config.ApplicationName())
	equals(t, "test-project", config.Project())
	equals(t, "abc123", config.TreeId())
	equals(t, "def789", config.ProjectEntropy())

	equals(t, "feature-x", config.Branch())
	equals(t, "feature-x-hgi456", config.Environment())
	equals(t, "/app/web", config.DocumentRoot())
	equals(t, "1.2.3.4", config.SmtpHost())
	equals(t, "8080", config.Port())
	equals(t, "unix://tmp/blah.sock", config.Socket())
}

func TestReadingExistingVariableWorks(t *testing.T) {
	config, err := NewConfigReal(runtimeEnv(envList{}), "PLATFORM_")
	ok(t, err)

	equals(t, "someval", config.Variable("somevar", ""))
}

func TestReadingMissingVariableReturnsDefault(t *testing.T) {
	config, err := NewConfigReal(runtimeEnv(envList{}), "PLATFORM_")
	ok(t, err)

	equals(t, "default-val", config.Variable("missing", "default-val"))
}

func TestVariablesReturnsMapWithData(t *testing.T) {
	config, err := NewConfigReal(runtimeEnv(envList{}), "PLATFORM_")
	ok(t, err)

	list := config.Variables()

	equals(t, "someval", list["somevar"])
}

func TestCredentialsForExistingRelationshipReturns(t *testing.T) {
	config, err := NewConfigReal(runtimeEnv(envList{}), "PLATFORM_")
	ok(t, err)

	creds, err := config.Credentials("database")
	ok(t, err)

	equals(t, "mysql", creds.Scheme)
}

//public function test_credentials_missing_relationship_throws() : void
func TestCredentialsForMissingRelationshipErrrors(t *testing.T) {
	config, err := NewConfigReal(runtimeEnv(envList{}), "PLATFORM_")
	ok(t, err)

	_, err = config.Credentials("does-not-exist")

	if err == nil {
		t.Fail()
	}
}

func TestGetAllRoutesAtRuntimeWorks(t *testing.T) {
	config, err := NewConfigReal(runtimeEnv(envList{}), "PLATFORM_")
	ok(t, err)

	routes, err := config.Routes()
	ok(t, err)

	equals(t, "upstream", routes["https://www.master-7rqtwti-gcpjkefjk4wc2.us-2.platformsh.site/"].Type)
}

func TestGetAllRoutesAtBuildtimeFails(t *testing.T) {
	config, err := NewConfigReal(buildEnv(envList{}), "PLATFORM_")
	ok(t, err)

	_, err = config.Routes()

	if err == nil {
		t.Fail()
	}
}

func TestGetRouteByIdWorks(t *testing.T) {
	config, err := NewConfigReal(runtimeEnv(envList{}), "PLATFORM_")
	ok(t, err)

	route, ok := config.Route("main")

	equals(t, true, ok)
	equals(t, "upstream", route.Type)
}

func TestGetNonExistentRouteErrors(t *testing.T) {
	config, err := NewConfigReal(runtimeEnv(envList{}), "PLATFORM_")
	ok(t, err)

	_, ok := config.Route("missing")

	equals(t, false, ok)
}

func TestSqlDsnIsFormattedCorrectly(t *testing.T) {
	config, err := NewConfigReal(runtimeEnv(envList{}), "PLATFORM_")
	ok(t, err)

	db, err := config.SqlDsn("database")
	ok(t, err)

	equals(t, "user:@tcp(database.internal:3306)/main?charset=utf8", db)

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
	env = mergeMaps(vars, env)
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
	env = mergeMaps(vars, env)
	env["PLATFORM_VARIABLES"] = encodeJsonFile("testdata/PLATFORM_VARIABLES.json")
	env["PLATFORM_APPLICATION"] = encodeJsonFile("testdata/PLATFORM_APPLICATION.json")
	env["PLATFORM_RELATIONSHIPS"] = encodeJsonFile("testdata/PLATFORM_RELATIONSHIPS.json")
	env["PLATFORM_ROUTES"] = encodeJsonFile("testdata/PLATFORM_ROUTES.json")

	vars = loadJsonFile("testdata/ENV_runtime.json")
	env = mergeMaps(vars, env)

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
