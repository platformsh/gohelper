// The gohelper library provides abstractions for the Platform.sh environment
// to make it easier to configure applications to run on Platform.sh.
// See https://docs.platform.sh/development/variables.html for an in-depth
// description of the available properties and their meaning.
package gohelper

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"os"
)

var NotValidPlatform = errors.New("No valid platform found.")

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

type Route struct {
	OriginalUrl    string            `json:"original_url"`
	Attributes     map[string]string `json:"attributes"`
	Type           string            `json:"type"`
	RestrictRobots bool              `json:"restrict_robots"`
	Tls            struct {
		ClientAuthentication         string   `json:"client_authentication"`
		MinVersion                   int      `json:"min_version"`
		ClientCertificateAuthorities []string `json:"client_certificate_authorities"`
		StrictTransportSecurity      struct {
			IncludeSubdomains bool `json:"include_subdomains"`
			Enabled           bool `json:"enabled"`
			Preload           bool `json:"preload"`
		}
	}
	Upstream string `json:"upstream"`
	Cache    struct {
		Enabled    bool     `json:"enabled"`
		Headers    []string `json:"headers"`
		Cookies    []string `json:"cookies"`
		DefaultTtl int      `json:"default_ttl"`
	}
	HttpAccess struct {
		Addresses []string          `json:"addresses"`
		BasicAuth map[string]string `json:"basic_auth"`
	}
	Primary bool   `json:"primary"`
	Id      string `json:"id"`
	Ssi     struct {
		Enabled bool `json:"enabled"`
	}

	// This field is not part of the JSON definition, but it gets added
	// to the struct from the JSON array key.
	Url string
}

type Routes map[string]Route

type PlatformConfig struct {
	// Prefixed simple values, build or deploy.
	applicationName string
	treeId          string
	appDir          string
	project         string
	projectEntropy  string

	// Prefixed simple values, runtime only.
	branch       string
	environment  string
	documentRoot string
	smtpHost     string
	mode         string

	// Prefixed complex values.
	credentials Credentials
	variables   envList
	routes      Routes
	application map[string]interface{}

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
		return nil, NotValidPlatform
	}

	// Extract the easy environment variables.
	p.applicationName = getter(p.prefix + "APPLICATION_NAME")
	p.appDir = getter(p.prefix + "APP_DIR")
	p.documentRoot = getter(p.prefix + "DOCUMENT_ROOT")
	p.treeId = getter(p.prefix + "TREE_ID")
	p.branch = getter(p.prefix + "BRANCH")
	p.environment = getter(p.prefix + "ENVIRONMENT")
	p.project = getter(p.prefix + "PROJECT")
	p.projectEntropy = getter(p.prefix + "PROJECT_ENTROPY")
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

	// Extract PLATFORM_ROUTES.
	if routes := getter(p.prefix + "ROUTES"); routes != "" {
		parsedRoutes, err := extractRoutes(routes)
		if err != nil {
			return nil, err
		}
		p.routes = parsedRoutes
	}

	// Extract PLATFORM_APPLICATION.
	// @todo Turn this into a proper struct.
	var parsedApplication map[string]interface{}
	jsonApplication, err := base64.StdEncoding.DecodeString(getter(p.prefix + "APPLICATION"))
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(jsonApplication, &parsedApplication)
	if err != nil {
		return nil, err
	}
	p.application = parsedApplication

	return p, nil
}

// This function returns a new PlatformConfig object, representing
// the abstracted Platform.sh environment.  If run on not a Platform.sh
// environment (eg, a local computer) then it will return nil and an error.
func NewConfig() (*PlatformConfig, error) {
	return NewConfigReal(os.Getenv, "PLATFORM_")
}

// Checks whether the code is running in a build environment.
func (p *PlatformConfig) InBuild() bool {
	return p.environment == ""
}

// Checks whether the code is running in a runtime environment.
func (p *PlatformConfig) InRuntime() bool {
	return p.environment != ""
}

// Determines if the current environment is a Platform.sh Enterprise environment.
func (p *PlatformConfig) OnEnterprise() bool {
	return p.mode == "enterprise"
}

// Determines if the current environment is a production environment.
//
// Note: There may be a few edge cases where this is not entirely correct on Enterprise,
// if the production branch is not named `production`.  In that case you'll need to use
// your own logic.
func (p *PlatformConfig) OnProduction() bool {
	if !p.InRuntime() {
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

// The name of the application, as defined in its configuration.
func (p *PlatformConfig) ApplicationName() string {
	return p.applicationName
}

// An ID identifying the application tree before it was built: a unique hash
// is generated based on the contents of the application's files in the
// repository.
func (p *PlatformConfig) TreeId() string {
	return p.treeId
}

// The absolute path to the application.
func (p *PlatformConfig) AppDir() string {
	return p.appDir
}

// The project ID.
func (p *PlatformConfig) Project() string {
	return p.project
}

// A random string generated for each project, useful for generating hash keys.
func (p *PlatformConfig) ProjectEntropy() string {
	return p.projectEntropy
}

// The Git branch name.
func (p *PlatformConfig) Branch() string {
	return p.branch
}

// The environment ID (usually the Git branch plus a hash).
func (p *PlatformConfig) Environment() string {
	return p.environment
}

// The absolute path to the web root of the application.
func (p *PlatformConfig) DocumentRoot() string {
	return p.documentRoot
}

// The hostname of the Platform.sh default SMTP server (an empty string if
// emails are disabled on the environment).
func (p *PlatformConfig) SmtpHost() string {
	return p.smtpHost
}

// The TCP port number the application should listen to for incoming requests.
func (p *PlatformConfig) Port() string {
	return p.port
}

// The Unix socket the application should listen to for incoming requests.
func (p *PlatformConfig) Socket() string {
	return p.socket
}

// Returns a variable from the VARIABLES array.
//
// Note: variables prefixed with `env:` can be accessed as normal environment variables.
// This method will return such a variable by the name with the prefix still included.
// Generally it's better to access those variables directly.
func (p *PlatformConfig) Variable(name string, defaultValue string) string {
	if val, ok := p.variables[name]; ok {
		return val
	}
	return defaultValue
}

// Returns the full variables array.
//
// If you're looking for a specific variable, the Variable() method is a more robust option.
// This method is for cases where you want to scan the whole variables list looking for a pattern.
func (p *PlatformConfig) Variables() envList {
	return p.variables
}

// Retrieves the credentials for accessing a relationship.
func (p *PlatformConfig) Credentials(relationship string) (Credential, error) {

	// Non-zero relationship indexes are not currently used, so hard code 0 for now.
	// On the off chance that ever changes, we'll add another method that allows
	// callers to specify an offset.
	if creds, ok := p.credentials[relationship]; ok {
		return creds[0], nil
	}

	return Credential{}, fmt.Errorf("No such relationship: %s", relationship)
}

// Returns the routes definition.
// This is an slice of Route structs.
func (p *PlatformConfig) Routes() (Routes, error) {
	if p.InBuild() {
		return Routes{}, fmt.Errorf("Routes are not available during the build phase.")
	}

	return p.routes, nil
}

// Returns a single route definition.
//
// Note: If no route ID was specified in routes.yaml then it will not be possible
// to look up a route by ID.
func (p *PlatformConfig) Route(id string) (Route, bool) {
	for _, route := range p.routes {
		if route.Id == id {
			return route, true
		}
	}

	return Route{}, false
}

// SqlDsn produces an SQL connection string appropriate for use with many
// common Go database tools.  If the relationship specified is not found
// or is not an SQL connection an error will be returned.
func (p *PlatformConfig) SqlDsn(name string) (string, error) {
	creds, err := p.Credentials(name)
	if err != nil {
		return "", err
	}

	dbString := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8", creds.Username, creds.Password, creds.Host, creds.Port, creds.Path)
	return dbString, nil
}

// Map the relationships environment variable string into the appropriate data structure.
func extractCredentials(relationships string) (Credentials, error) {
	jsonRelationships, err := base64.StdEncoding.DecodeString(relationships)
	if err != nil {
		return Credentials{}, err
	}

	var rels Credentials

	err = json.Unmarshal([]byte(jsonRelationships), &rels)
	if err != nil {
		return nil, err
	}

	return rels, nil
}

// Map the variables environment variable string into the appropriate data structure.
func extractVariables(vars string) (envList, error) {
	jsonVars, err := base64.StdEncoding.DecodeString(vars)
	if err != nil {
		return envList{}, err
	}

	var env envList

	err = json.Unmarshal([]byte(jsonVars), &env)
	if err != nil {
		return nil, err
	}

	return env, nil
}

// Map the routes environment variable string into the appropriate data structure.
func extractRoutes(routesString string) (Routes, error) {
	jsonRoutes, err := base64.StdEncoding.DecodeString(routesString)
	if err != nil {
		return Routes{}, err
	}

	var routes Routes

	err = json.Unmarshal([]byte(jsonRoutes), &routes)
	if err != nil {
		return nil, err
	}

	// Normalize the URL of each route into the struct, so that it's available
	// when requesting a route individually.
	for url, route := range routes {
		route.Url = url
	}

	return routes, nil
}
