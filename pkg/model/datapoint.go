package model

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
	Interval_   PollInterval `json:"interval"`
	Unit_       string       `json:"unit"` // Allow all characters
	SourceType_ SourceType   `json:"sourcetype"`

	Scaling    Scaling    `json:"scaling"`
	K          float64    `json:"k"`
	M          float64    `json:"m"`
	TimeToLive TimeToLive `json:"timeToLive"`
}

type Ttnv3Document struct {
	DatapointSettings

	Zone             string `json:"zone"`
	Application      string `json:"application"`
	Device           string `json:"device"`
	PointName        string `json:"pointname"`
	AuthorizationKey string `json:"authorizationkey"`
}

type WebDocument struct {
	DatapointSettings

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

func (w *WebDocument) SourceType() SourceType {
	return Web
}

func (w *WebDocument) Project() string {
	return w.Project_
}

func (w *WebDocument) Subsystem() string {
	return w.Subsystem_
}

func (w *WebDocument) Name() string {
	return w.Name_
}

func (w *WebDocument) Unit() string {
	return w.Unit_
}

func (w *WebDocument) Interval() PollInterval {
	return w.Interval_
}

func (t *Ttnv3Document) SourceType() SourceType {
	return Ttnv3
}

func (t *Ttnv3Document) Project() string {
	return t.Project_
}

func (t *Ttnv3Document) Subsystem() string {
	return t.Subsystem_
}

func (t *Ttnv3Document) Name() string {
	return t.Name_
}

func (t *Ttnv3Document) Unit() string {
	return t.Unit_
}
func (t *Ttnv3Document) Interval() PollInterval {
	return t.Interval_
}

type SourceType string

const (
	Web   SourceType = "web"   // Web documents
	Ttnv3 SourceType = "ttnv3" // The Things Network v3
)
