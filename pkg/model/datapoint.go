package model

import (
	"github.com/gocql/gocql"
	"strconv"
)

type Datapoint interface {
	SourceType() SourceType
	Project() string
	Subsystem() string
	Name() string
	Unit() string
	Interval() PollInterval
}

type DatapointSettings struct {
	Project    string       `json:"project"`
	Subsystem  string       `json:"subsystem"`
	Name       string       `json:"name"` // validate regexp [a-z][A-Za-z0-9_.]*
	Interval   PollInterval `json:"pollinterval"`
	Proc       Processing   `json:"proc"`
	TimeToLive TimeToLive   `json:"timeToLive"`
	SourceType SourceType   `json:"datasourcetype"`
	Datasource interface{}  `json:"datasource"` // either a Ttnv3Datasource or a WebDatasource depending on SourceType
}

type Processing struct {
	Unit      string  `json:"unit"` // Allow all characters
	Scaling   Scaling `json:"scaling"`
	K         float64 `json:"k"`
	M         float64 `json:"m"`
	Min       float64 `json:"min"`
	Max       float64 `json:"max"`
	Condition string  `json:"condition"` // Allow all characters
	ScaleFunc string  `json:"scalefunc"` // Allow all characters
}

func (p *Processing) UnmarshalUDT(name string, info gocql.TypeInfo, data []byte) error {
	switch name {
	case "unit":
		d := string(data)
		p.Unit = d
	case "scaling":
		d := string(data)
		switch {
		case d == "lin":
			p.Scaling = Lin
		case d == "ln":
			p.Scaling = Ln
		case d == "exp":
			p.Scaling = Exp
		case d == "rad":
			p.Scaling = Rad
		case d == "deg":
			p.Scaling = Deg
		case d == "fToC":
			p.Scaling = FtoC
		case d == "cToF":
			p.Scaling = CtoF
		case d == "kToC":
			p.Scaling = KtoC
		case d == "cToK":
			p.Scaling = CtoK
		case d == "kToF":
			p.Scaling = KtoF
		case d == "fToK":
			p.Scaling = FtoK
		}
	case "k":
		d := string(data)
		f, err := strconv.ParseFloat(d, 64)
		if err == nil {
			p.K = f
		}
	case "m":
		d := string(data)
		f, err := strconv.ParseFloat(d, 64)
		if err == nil {
			p.M = f
		}
	case "min":
		d := string(data)
		f, err := strconv.ParseFloat(d, 64)
		if err == nil {
			p.Min = f
		}
	case "max":
		d := string(data)
		f, err := strconv.ParseFloat(d, 64)
		if err == nil {
			p.Max = f
		}
	case "condition":
		d := string(data)
		p.Condition = d
	case "scalefunc":
		d := string(data)
		p.ScaleFunc = d
	}
	return nil
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
		ds.Zone = string(data)
	case "application":
		ds.Application = string(data)
	case "device":
		ds.Device = string(data)
	case "pointname":
		ds.PointName = string(data)
	case "authorizationkey":
		ds.AuthorizationKey = string(data)
	}
	return nil
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
		ds.URL = string(data)
	case "authtype":
		t := string(data)
		switch {
		case t == "none":
			ds.AuthenticationType = None
		case t == "basic":
			ds.AuthenticationType = Basic
		case t == "bearerToken":
			ds.AuthenticationType = BearerToken
		}
	case "auth":
		ds.Auth = string(data)
	case "doctype":
		t := string(data)
		switch {
		case t == "jsondoc":
			ds.Format = JSON
		case t == "xmldoc":
			ds.Format = XML
		}
	case "dataexpr":
		ds.ValueExpression = string(data)
	case "tstype":
		t := string(data)
		switch {
		case t == "polltime":
			ds.TimestampType = PollTime
		case t == "epochMillis":
			ds.TimestampType = EpochMillis
		case t == "epochSeconds":
			ds.TimestampType = EpochSeconds
		case t == "iso8601_zoned":
			ds.TimestampType = ISO8601_zoned
		case t == "iso8601_offset":
			ds.TimestampType = ISO8601_offset
		}
	case "tsexpr":
		ds.TimestampExpression = string(data)
	}
	return nil
}

type SourceType string

const (
	Web   SourceType = "web"   // Web documents
	Ttnv3 SourceType = "ttnv3" // The Things Network v3
)
