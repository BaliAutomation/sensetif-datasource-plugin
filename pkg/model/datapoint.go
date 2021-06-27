package model

type DatapointSettings struct {
	Project   string       `json:"project"`
	Subsystem string       `json:"subsystem"`
	Name      string       `json:"name"` // validate regexp [a-z][A-Za-z0-9_.]*
	Interval  PollInterval `json:"interval"`
	URL       string       `json:"url"` // validate URL, incl anchor and query arguments, but disallow user pwd@

	// Authentication is going to need a lot in the future, but for now user/pass is fine
	AuthenticationType AuthenticationType `json:"authenticationType"`
	Auth               map[string]string  `json:"auth"` // if Authentication == basic, then map contains keys "u" and "p"

	Format          OriginDocumentFormat `json:"format"`
	ValueExpression string               `json:"valueExpression"` // if format==xml, then xpath. if format==json, then jsonpath. If there is library available for validation, do that. If not, put in a function and we figure that out later.
	Unit            string               `json:"unit"`            // Allow all characters

	// Ideally only show k and m for ScalingFunctions that uses them, and show the formula with the scaling function
	Scaling Scaling `json:"scaling"`
	K       float64 `json:"k"`
	M       float64 `json:"m"`

	TimestampType       TimestampType `json:"timestampType"`
	TimestampExpression string        `json:"timestampExpression"` // if format==xml, then xpath. if format==json, then jsonpath.
	TimeToLive          TimeToLive    `json:"timeToLive"`
}
