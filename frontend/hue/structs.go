package hue

type LightState struct {
	On        bool      `json:"on"`
	Bri       int       `json:"bri"`
	Hue       int       `json:"hue"`
	Sat       int       `json:"sat"`
	Effect    string    `json:"effect"`
	Ct        int       `json:"ct"`
	Alert     string    `json:"alert"`
	Colormode string    `json:"colormode"`
	Reachable bool      `json:"reachable"`
	XY        []float64 `json:"xy"`
}

type LightStateChange struct {
	On             *bool   `json:"on,omitempty"`
	Bri            *int    `json:"bri,omitempty"`
	Hue            *int    `json:"hue,omitempty"`
	Sat            *int    `json:"sat,omitempty"`
	Effect         *string `json:"effect,omitempty"`
	Ct             *int    `json:"ct,omitempty"`
	Alert          *string `json:"alert,omitempty"`
	Colormode      *string `json:"colormode,omitempty"`
	TransitionTime int     `json:"transitiontime,omitempty"`
}

type LightStateChangeResponse []struct {
	Success map[string]interface{} `json:"success,omitempty"`
}

type Light struct {
	State            LightState `json:"state"`
	Type             string     `json:"type"`
	Name             string     `json:"name"`
	ModelID          string     `json:"modelid"`
	ManufacturerName string     `json:"manufacturername"`
	UniqueID         string     `json:"uniqueid"`
	SwVersion        string     `json:"swversion"`
	PointSymbol      struct {
		One   string `json:"1"`
		Two   string `json:"2"`
		Three string `json:"3"`
		Four  string `json:"4"`
		Five  string `json:"5"`
		Six   string `json:"6"`
		Seven string `json:"7"`
		Eight string `json:"8"`
	} `json:"pointsymbol"`
}

type LightList map[string]Light
