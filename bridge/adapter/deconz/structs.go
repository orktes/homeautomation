package deconz

import (
	"bytes"
	"encoding/json"
	"errors"
	"time"
)

type customTime struct {
	time.Time
}

func (t customTime) MarshalJSON() ([]byte, error) {
	if y := t.Year(); y < 0 || y >= 10000 {
		// RFC 3339 is clear that years are 4 digits exactly.
		// See golang.org/issue/4556#c15 for more discussion.
		return nil, errors.New("Time.MarshalJSON: year outside of range [0,9999]")
	}

	b := make([]byte, 0, len(time.RFC3339Nano)+2)
	b = append(b, '"')
	b = t.AppendFormat(b, time.RFC3339Nano)
	b = append(b, '"')
	return b, nil
}

// UnmarshalJSON implements the json.Unmarshaler interface.
// The time is expected to be a quoted string in RFC 3339 format.
func (t *customTime) UnmarshalJSON(data []byte) error {
	// Ignore null, like in the main JSON package.
	if string(data) == "null" {
		return nil
	}

	adjust := false
	if bytes.Index(data, []byte("Z")) == -1 {
		// TODO get correct time offset
		data = append(data[:len(data)-1], []byte("Z\"")...)
		adjust = true
	}

	// Fractional seconds are handled implicitly by Parse.
	var err error
	t.Time, err = time.Parse(`"`+time.RFC3339+`"`, string(data))

	if adjust {
		t := time.Now()
		_, offset := t.Zone()
		t.Add(time.Duration(offset) * time.Second)
		// TODO might not be the right thing always if deconz is not running on the same server
	}
	return err
}

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
	ButtonEvent *int        `json:"buttonevent"`
	LastUpdated *customTime `json:"lastupdated"`
	Dark        *bool       `json:"dark"`
	Daylight    *bool       `json:"daylight"`
	LightLevel  *int        `json:"lightlevel"`
	Lux         *int        `json:"lux"`
	Presence    *bool       `json:"presence"`
}

type sensor struct {
	Name  string      `json:"name"`
	State sensorState `json:"state"`
}
