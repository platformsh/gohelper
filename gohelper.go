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
