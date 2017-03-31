package gohelper

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
)

type RelationshipQuery struct {
	IsMaster bool `json:"is_master"`
}

type Relationship struct {
	Host     string `json:"host"`
	Username string `json:"username"`
	Password string `json:"password"`
	Ip       string `json:"ip"`
	Path     string `json:"path"`
	Scheme   string `json:"scheme"`
	Port     int    `json:"port"`
	Query    RelationshipQuery
}
type Relationships map[string][]Relationship

func (p *PlatformInfo) SqlDsn(name string) (string, error) {

	//fmt.Printf("%+v\n", p.Relationships[name])

	if relInfo, ok := p.Relationships[name]; ok {

		if len(relInfo) > 0 {
			dbInfo := relInfo[0]
			dbString := fmt.Sprintf("%s:%s@%s:%s/%s?charset=utf8", dbInfo.Username, dbInfo.Password, dbInfo.Host, dbInfo.Port, dbInfo.Path)
			fmt.Println(dbString)
			return dbString, nil
		}
		return "", fmt.Errorf("No first relationship defined for: %s.", name)
	}

	return "", fmt.Errorf("No such relationship defined: %s.", name)
}

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

func getPlatformshRelationships() (Relationships, error) {

	relationships := os.Getenv("PLATFORM_RELATIONSHIPS")

	//fmt.Println(relationships)

	jsonRelationships, _ := base64.StdEncoding.DecodeString(relationships)

	//fmt.Println(string(sDec))

	var rels Relationships

	//fmt.Println("A")

	err := json.Unmarshal([]byte(jsonRelationships), &rels)
	if err != nil {
		//fmt.Println("B")
		return nil, err
	}

	//fmt.Println("C")

	//fmt.Printf("%+v\n", rels)
	//fmt.Printf("%+v\n", rels["mysql"][0])

	//fmt.Println(rels["mysql"][0].Host)

	return rels, nil
}
