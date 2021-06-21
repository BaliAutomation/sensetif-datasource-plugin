package model

type ProjectSettings struct {
	Name        string `json:"name"`        // validate regexp[a-z][A-Za-z0-9_]*
	Title       string `json:"title"`       // allow all characters
	City        string `json:"city"`        // allow all characters
	Country     string `json:"country"`     // country list?
	Timezone    string `json:"timezone"`    // "UTC" or {continent}/{city}, ex Europe/Stockholm
	Geolocation string `json:"geolocation"` // geo coordinates
}
