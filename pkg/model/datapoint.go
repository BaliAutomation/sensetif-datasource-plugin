package model

import "github.com/gocql/gocql"

type Datapoint interface {
	SourceType() SourceType
	Project() string
	Subsystem() string
	Name() string
	Unit() string
	Interval() PollInterval
}

type DatapointSettings struct {
	Project_    string       `json:"project"`
	Subsystem_  string       `json:"subsystem"`
	Name_       string       `json:"name"` // validate regexp [a-z][A-Za-z0-9_.]*
	Interval_   PollInterval `json:"pollinterval"`
	Proc        Processing   `json:"proc"`
	TimeToLive  TimeToLive   `json:"timeToLive"`
	SourceType_ SourceType   `json:"datasourcetype"`
	Datasource  interface{}  `json:"datasource"` // either a Ttnv3Datasource or a WebDatasource depending on SourceType
}

type Processing struct {
	Unit_      string  `json:"unit"` // Allow all characters
	Scaling    Scaling `json:"scaling"`
	K          float64 `json:"k"`
	M          float64 `json:"m"`
	Min        float64 `json:"min"`
	Max        float64 `json:"max"`
	Condition_ string  `json:"condition"` // Allow all characters
	ScaleFunc  string  `json:"scalefunc"` // Allow all characters
}

func (p *Processing) UnmarshalUDT(name string, info gocql.TypeInfo, data []byte) error {
	switch name {
	case "unit":
		return gocql.Unmarshal(info, data, &p.Unit_)
	case "condition":
		return gocql.Unmarshal(info, data, &p.Condition_)
	case "scalefunc":
		return gocql.Unmarshal(info, data, &p.ScaleFunc)
	case "scaling":
		return gocql.Unmarshal(info, data, &p.Scaling)
	case "k":
		return gocql.Unmarshal(info, data, &p.K)
	case "m":
		return gocql.Unmarshal(info, data, &p.M)
	case "min":
		return gocql.Unmarshal(info, data, &p.Min)
	case "max":
		return gocql.Unmarshal(info, data, &p.Max)
	default:
		return nil
	}
}

type Ttnv3Datasource struct {
	Zone             string `json:"zone"`
	Application      string `json:"application"`
	Device           string `json:"device"`
	PointName        string `json:"pointname"`
	AuthorizationKey string `json:"authorizationkey"`
}

func (ds *Ttnv3Datasource) UnmarshalUDT(name string, info gocql.TypeInfo, data []byte) error {
	switch name {
	case "zone":
		return gocql.Unmarshal(info, data, &ds.Zone)
	case "application":
		return gocql.Unmarshal(info, data, &ds.Application)
	case "device":
		return gocql.Unmarshal(info, data, &ds.Device)
	case "pointname":
		return gocql.Unmarshal(info, data, &ds.PointName)
	case "authorizationkey":
		return gocql.Unmarshal(info, data, &ds.AuthorizationKey)
	default:
		return nil
	}
}

type WebDatasource struct {
	URL string `json:"url"` // validate URL, incl anchor and query arguments, but disallow user pwd@

	// Authentication is going to need a lot in the future, but for now user/pass is fine
	AuthenticationType AuthenticationType `json:"authenticationType"`

	// if Authentication == basic, then string contains [user]"="[password]
	// if Authentication == bearerToken then string contains token wihout "Bearer"
	Auth string `json:"auth"`

	Format              OriginDocumentFormat `json:"format"`
	ValueExpression     string               `json:"valueExpression"` // if format==xml, then xpath. if format==json, then jsonpath. If there is library available for validation, do that. If not, put in a function and we figure that out later.
	TimestampType       TimestampType        `json:"timestampType"`
	TimestampExpression string               `json:"timestampExpression"` // if format==xml, then xpath. if format==json, then jsonpath.
}

func (ds *WebDatasource) UnmarshalUDT(name string, info gocql.TypeInfo, data []byte) error {
	switch name {
	case "url":
		return gocql.Unmarshal(info, data, &ds.URL)
	case "authenticationType":
		return gocql.Unmarshal(info, data, &ds.AuthenticationType)
	case "auth":
		return gocql.Unmarshal(info, data, &ds.Auth)
	case "format":
		return gocql.Unmarshal(info, data, &ds.Format)
	case "valueExpression":
		return gocql.Unmarshal(info, data, &ds.ValueExpression)
	case "timestampType":
		return gocql.Unmarshal(info, data, &ds.TimestampType)
	case "timestampExpression":
		return gocql.Unmarshal(info, data, &ds.TimestampExpression)
	default:
		return nil
	}
}

type SourceType string

const (
	Web   SourceType = "web"   // Web documents
	Ttnv3 SourceType = "ttnv3" // The Things Network v3
)
