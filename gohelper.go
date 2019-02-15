// The gohelper library provides abstractions for the Platform.sh environment
// to make it easier to configure applications to run on Platform.sh.
// See https://docs.platform.sh/development/variables.html for an in-depth
// description of the available properties and their meaning.
package gohelper

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	//"github.com/davecgh/go-spew/spew"

	//"fmt"
	"os"
)

type Error string

func (e Error) Error() string { return string(e) }

const notValidPlatform = Error("No valid platform found.")

type envList map[string]string

type envReader func(string) string

type Credential struct {
	Scheme   string `json:"scheme"`
	Cluster  string `json:"cluster"`
	Service  string `json:"service"`
	Username string `json:"username"`
	Password string `json:"password"`
	Host     string `json:"host"`
	Path     string `json:"path"`
	Public   bool   `json:"public"`
	Fragment string `json:"fragment"`
	Ip       string `json:"ip"`
	Rel      string `json:"rel"`
	Type     string `json:"type"`
	Port     int    `json:"port"`
	Hostname string `json:"hostname"`
	Query    struct {
		IsMaster bool `json:"is_master"`
	}
}

type Credentials map[string][]Credential

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
	credentials Credentials
	variables   envList
	//Application     ApplicationInfo
	//Routes          RouteInfo

	// Unprefixed simple values.
	socket string
	port   string

	// Internal data.
	prefix string
}

func NewConfigReal(getter envReader, prefix string) (*PlatformConfig, error) {
	p := &PlatformConfig{}

	p.prefix = prefix

	// If it's not a valid platform, bail out now.
	if getter(prefix+"APPLICATION_NAME") == "" {
		return nil, notValidPlatform
	}

	// Extract the easy environment variables.
	p.applicationName = getter(p.prefix + "APPLICATION_NAME")
	p.appDir = getter(p.prefix + "APP_DIR")
	p.documentRoot = getter(p.prefix + "DOCUMENT_ROOT")
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

	// Extract PLATFORM_RELATIONSHIPS, which we'll call credentials since that's what they are.
	if rels := getter(p.prefix + "RELATIONSHIPS"); rels != "" {
		creds, err := extractCredentials(rels)
		if err != nil {
			return nil, err
		}
		p.credentials = creds
	}

	// Extract the PLATFORM_VARIABLES array.
	if vars := getter(p.prefix + "VARIABLES"); vars != "" {
		parsedVars, err := extractVariables(vars)
		if err != nil {
			return nil, err
		}
		p.variables = parsedVars
	}

	// @todo extract PLATFORM_ROUTES

	// @todo extract PLATFORM_APPLICATION (oh dear oh dear)

	return p, nil
}

func NewConfig() (*PlatformConfig, error) {
	return NewConfigReal(os.Getenv, "PLATFORM_")
}

func (p *PlatformConfig) InBuild() bool {
	return p.environment == ""
}

func (p *PlatformConfig) OnEnterprise() bool {
	return p.mode == "enterprise"
}

func (p *PlatformConfig) OnProduction() bool {
	if p.InBuild() {
		return false
	}

	var prodBranch string
	if p.OnEnterprise() {
		prodBranch = "production"
	} else {
		prodBranch = "master"
	}

	return p.branch == prodBranch
}

func (p *PlatformConfig) ApplicationName() string {
	return p.applicationName
}

func (p *PlatformConfig) TreeId() string {
	return p.treeId
}

func (p *PlatformConfig) AppDir() string {
	return p.appDir
}

func (p *PlatformConfig) Project() string {
	return p.project
}

func (p *PlatformConfig) Entropy() string {
	return p.entropy
}

func (p *PlatformConfig) Branch() string {
	return p.branch
}

func (p *PlatformConfig) Environment() string {
	return p.environment
}

func (p *PlatformConfig) DocumentRoot() string {
	return p.documentRoot
}

func (p *PlatformConfig) SmtpHost() string {
	return p.smtpHost
}

func (p *PlatformConfig) Port() string {
	return p.port
}

func (p *PlatformConfig) Socket() string {
	return p.socket
}

func (p *PlatformConfig) Variable(name string, defaultValue string) string {
	if val, ok := p.variables[name]; ok {
		return val
	}
	return defaultValue
}

func (p *PlatformConfig) Variables() envList {
	return p.variables
}

func (p *PlatformConfig) Credentials(relationship string) (Credential, error) {

	// Non-zero relationship indexes are not currently used, so hard code 0 for now.
	// On the off chance that ever changes, we'll add another method that allows
	// callers to specify an offset.
	if creds, ok := p.credentials[relationship]; ok {
		return creds[0], nil
	}

	return Credential{}, fmt.Errorf("No such relationship: %s", relationship)
}

// SqlDsn produces an SQL connection string appropriate for use with many
// common Go database tools.  If the relationship specified is not found
// or is not an SQL connection an error will be returned.
/*
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
*/

func extractCredentials(relationships string) (Credentials, error) {
	jsonRelationships, _ := base64.StdEncoding.DecodeString(relationships)

	var rels Credentials

	err := json.Unmarshal([]byte(jsonRelationships), &rels)
	if err != nil {
		return nil, err
	}

	return rels, nil
}

func extractVariables(vars string) (envList, error) {
	jsonVars, _ := base64.StdEncoding.DecodeString(vars)

	var env envList

	err := json.Unmarshal([]byte(jsonVars), &env)
	if err != nil {
		return nil, err
	}

	return env, nil
}
