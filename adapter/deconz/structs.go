package deconz

import "encoding/json"

type configResponse struct {
	WebsocketPort int    `json:"websocketport"`
	IPAddress     string `json:"ipaddress"`
}

type event struct {
	Event string          `json:"e"`
	ID    string          `json:"id"`
	Route string          `json:"r"`
	State json.RawMessage `json:"state"`
	Type  string          `json:"t"`
}

func (ev *event) unmarshalState(in interface{}) error {
	return json.Unmarshal(ev.State, in)
}

type lightStateChangeResponse []struct {
	Success map[string]interface{} `json:"success,omitempty"`
}

type lightState struct {
	On        *bool      `json:"on"`
	Bri       *int       `json:"bri"`
	Hue       *int       `json:"hue"`
	Sat       *int       `json:"sat"`
	Effect    *string    `json:"effect"`
	Ct        *int       `json:"ct"`
	Alert     *string    `json:"alert"`
	Colormode *string    `json:"colormode"`
	Reachable *bool      `json:"reachable"`
	XY        *[]float64 `json:"xy"`
}

type light struct {
	State            lightState `json:"state"`
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

type groupState struct {
	AnyOn *bool `json:"any_on"`
}

type group struct {
	Action lightState `json:"action"`
	Name   string     `json:"name"`
	Lights []string   `json:"lights"`
	Scenes []struct {
		ID         string `json:"id"`
		Name       string `json:"name"`
		LightCount int    `json:"lightcount"`
	}
	State groupState `json:"state"`
}

type sensorState struct {
	ButtonEvent *int  `json:"buttonevent"`
	Dark        *bool `json:"dark"`
	Daylight    *bool `json:"daylight"`
	LightLevel  *int  `json:"lightlevel"`
	Lux         *int  `json:"lux"`
	Presence    *bool `json:"presence"`
}

type sensor struct {
	State sensorState `json:"state"`
}
