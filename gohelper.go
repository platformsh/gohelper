// The gohelper library provides abstractions for the Platform.sh environment
// to make it easier to configure applications to run on Platform.sh.
// See https://docs.platform.sh/development/variables.html for an in-depth
// description of the available properties and their meaning.
package gohelper

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
)

type Error string

func (e Error) Error() string { return string(e) }

const notValidPlatform = Error("No valid platform found.")

type Relationship struct {
	Host     string `json:"host"`
	Username string `json:"username"`
	Password string `json:"password"`
	Ip       string `json:"ip"`
	Path     string `json:"path"`
	Scheme   string `json:"scheme"`
	Port     int    `json:"port"`
	Query    struct {
		IsMaster bool `json:"is_master"`
	}
}
type Relationships map[string][]Relationship

type PlatformInfo struct {
	Relationships Relationships
	//Application     ApplicationInfo
	//Routes          RouteInfo
	//Variables       map[string]string
	ApplicationName string
	DocRoot         string
	Branch          string
	TreeId          string
	AppDir          string
	Environment     string
	Project         string
	Entropy         string
	Socket          string
	Port            string
}

type PlatformConfig struct {
	// Prefixed simple values, build or deploy.
	applicationName string
	treeId          string
	appDir          string
	project         string
	entropy         string

	// Prefixed simple values, runtime only.
	branch       string
	environment  string
	documentRoot string
	smtpHost     string
	mode         string

	// Prefixed complex values.
	relationships Relationships
	//Application     ApplicationInfo
	//Routes          RouteInfo
	//Variables       map[string]string

	// Unprefixed simple values.
	socket string
	port   string

	// Internal data.
	prefix string
}

func NewConfigReal(getter func(string) string) (*PlatformConfig, error) {
	p := &PlatformConfig{}

	switch {
	case getter("PLATFORM_APPLICATION_NAME") != "":
		p.prefix = "PLATFORM_"
	case getter("MAGENTO_APPLICATION_NAME") != "":
		p.prefix = "MAGENTO_"
	case getter("SYMFONY_APPLICATION_NAME") != "":
		p.prefix = "SYMFONY_"
	default:
		return nil, notValidPlatform
	}

	// Extract the easy environment variables.
	p.applicationName = getter(p.prefix + "APPLICATION_NAME")
	p.appDir = getter(p.prefix + "APP_DIR")
	p.documentRoot = getter(p.prefix + "_DOCUMENT_ROOT")
	p.treeId = getter(p.prefix + "TREE_ID")
	p.branch = getter(p.prefix + "BRANCH")
	p.environment = getter(p.prefix + "ENVIRONMENT")
	p.project = getter(p.prefix + "PROJECT")
	p.entropy = getter(p.prefix + "PROJECT_ENTROPY")
	p.smtpHost = getter(p.prefix + "SMTP_HOST")
	p.mode = getter(p.prefix + "MODE")
	p.socket = getter("SOCKET")
	p.port = getter("PORT")

	// Extract the complex environment variables (serialized JSON strings).
	// @todo Rename this to credentials, at least externally.
	/*
		rels, err := getPlatformshRelationships()
		if err != nil {
			return nil, err
		}
		p.relationships = rels
	*/

	// @todo extract the PLATFORM_VARIABLES array

	// @todo extract PLATFORM_ROUTES

	// @todo extract PLATFORM_APPLICATION (oh dear oh dear)

	return p, nil
}

func NewConfig() (*PlatformConfig, error) {
	return NewConfigReal(os.Getenv)
}

func (p *PlatformConfig) InBuild() bool {
	return p.environment == ""
}

// NewPlatformInfo returns a struct containing environment information
// for the current Platform.sh environment. That includes the port on
// which to listen for web requests, database credentials, and so on.
// If that information is not available due to being called when not
// running on Platform.sh an error will be returned.
func NewPlatformInfo() (*PlatformInfo, error) {
	p := &PlatformInfo{}

	// Extract the complex environment variables (serialized JSON strings).
	rels, err := getPlatformshRelationships()
	if err != nil {
		return nil, err
	}
	p.Relationships = rels

	// Extract the easy stuff.
	p.ApplicationName = os.Getenv("PLATFORM_APPLICATION_NAME")
	p.AppDir = os.Getenv("PLATFORM_APP_DIR")
	p.DocRoot = os.Getenv("PLATFORM_DOCUMENT_ROOT")
	p.TreeId = os.Getenv("PLATFORM_TREE_ID")
	p.Branch = os.Getenv("PLATFORM_BRANCH")
	p.Environment = os.Getenv("PLATFORM_ENVIRONMENT")
	p.Project = os.Getenv("PLATFORM_PROJECT")
	p.Entropy = os.Getenv("PLATFORM_PROJECT_ENTROPY")
	p.Socket = os.Getenv("SOCKET")
	p.Port = os.Getenv("PORT")

	return p, nil
}

// SqlDsn produces an SQL connection string appropriate for use with many
// common Go database tools.  If the relationship specified is not found
// or is not an SQL connection an error will be returned.
func (p *PlatformInfo) SqlDsn(name string) (string, error) {
	if relInfo, ok := p.Relationships[name]; ok {
		if len(relInfo) > 0 {
			dbInfo := relInfo[0]
			dbString := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8", dbInfo.Username, dbInfo.Password, dbInfo.Host, dbInfo.Port, dbInfo.Path)
			return dbString, nil
		}
		return "", fmt.Errorf("No first relationship defined for: %s.", name)
	}

	return "", fmt.Errorf("No such relationship defined: %s.", name)
}

func getPlatformshRelationships() (Relationships, error) {

	relationships := os.Getenv("PLATFORM_RELATIONSHIPS")
	jsonRelationships, _ := base64.StdEncoding.DecodeString(relationships)

	var rels Relationships

	err := json.Unmarshal([]byte(jsonRelationships), &rels)
	if err != nil {
		return nil, err
	}

	return rels, nil
}
