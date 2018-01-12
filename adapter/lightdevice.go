package adapter

// LightDevice interface that all lights should implement
type LightDevice interface {
	Device
	GetLightType() string
	GetLightState() LightState
	SetLightState(state LightState)
}

// LightState represents a light state
type LightState struct {
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
