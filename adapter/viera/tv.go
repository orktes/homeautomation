package viera

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"sync"
	"time"

	"text/template"

	"github.com/orktes/homeautomation/adapter"
	wol "github.com/sabhiram/go-wol"
)

var UPDATE_LOOP_INTERVAL = 5

var cmdTemplate = template.Must(template.New("viera_cmd").Parse(`<?xml version='1.0' encoding='utf-8'?>
<s:Envelope xmlns:s='http://schemas.xmlsoap.org/soap/envelope/' s:encodingStyle='http://schemas.xmlsoap.org/soap/encoding/'>
<s:Body>
<u:{{.action}} xmlns:u='urn:{{.urn}}'>
	{{.command}}
</u:{{.action}}>
</s:Body>
</s:Envelope>
`))

var currentVolumeRegex = regexp.MustCompile("<CurrentVolume>([0-9]+)</CurrentVolume>")
var currentMuteRegex = regexp.MustCompile("<CurrentMute>([0-9]+)</CurrentMute>")

type VieraTV struct {
	id string

	host string
	adapter.Updater
	mac string

	power  bool
	volume int
	mute   bool

	sync.Mutex
}

func (vt *VieraTV) init() error {
	if err := vt.readValues(false); err != nil {
		return err
	}

	go vt.updateLoop()

	return nil
}

func (vt *VieraTV) updateLoop() {
	for {
		time.Sleep(time.Duration(UPDATE_LOOP_INTERVAL) * time.Second)
		vt.readValues(true)
	}
}

func (vt *VieraTV) readValues(emitUpdate bool) error {
	if err := vt.readVolume(emitUpdate); err != nil {
		return err
	}

	if err := vt.readMute(emitUpdate); err != nil {
		return err
	}

	return nil
}

func (vt *VieraTV) setVolume(val int) error {
	vt.Lock()
	defer vt.Unlock()
	_, err := vt.sendCMD("render", "SetVolume", fmt.Sprintf("<InstanceID>0</InstanceID><Channel>Master</Channel><DesiredVolume>%d</DesiredVolume>", val))
	if val != vt.volume {
		vt.volume = val
		vt.SendUpdate(adapter.Update{
			ValueContainer: vt,
			Updates: []adapter.ValueUpdate{
				adapter.ValueUpdate{
					Key:   vt.id + ".volume",
					Value: val,
				},
			},
		})
	}

	return err
}

func (vt *VieraTV) setMute(val bool) error {
	vt.Lock()
	defer vt.Unlock()

	intVal := 0
	if val {
		intVal = 1
	}

	_, err := vt.sendCMD("render", "SetVolume", fmt.Sprintf("<InstanceID>0</InstanceID><Channel>Master</Channel><DesiredMute>%d</DesiredMute>", intVal))
	if val != vt.mute {
		vt.mute = val
		vt.SendUpdate(adapter.Update{
			ValueContainer: vt,
			Updates: []adapter.ValueUpdate{
				adapter.ValueUpdate{
					Key:   vt.id + ".mute",
					Value: val,
				},
			},
		})
	}

	return err
}

func (vt *VieraTV) setPower(power bool) error {
	vt.Lock()
	defer vt.Unlock()

	if power {
		wol.SendMagicPacket(vt.mac, "255.255.255.255:9", "")
	} else {
		if _, err := vt.sendCMD("control", "X_SendKey", "<X_KeyEvent>NRC_POWER-ONOFF</X_KeyEvent>"); err != nil {
			return err
		}
	}

	if vt.power == power {
		return nil
	}
	vt.power = power

	vt.SendUpdate(adapter.Update{
		ValueContainer: vt,
		Updates: []adapter.ValueUpdate{
			adapter.ValueUpdate{
				Key:   vt.id + ".power",
				Value: power,
			},
		},
	})

	return nil
}

func (vt *VieraTV) readVolume(emitUpdate bool) error {
	b, err := vt.sendCMD("render", "GetVolume", "<InstanceID>0</InstanceID><Channel>Master</Channel>")
	if err != nil {
		return err
	}

	vol := currentVolumeRegex.FindSubmatch(b)
	if len(vol) > 1 {
		intval, err := strconv.ParseInt(string(vol[1]), 10, 64)
		if err != nil {
			return err
		}

		vt.Lock()
		val := int(intval)
		if val != vt.volume {
			vt.volume = val
			if emitUpdate {
				vt.SendUpdate(adapter.Update{
					ValueContainer: vt,
					Updates: []adapter.ValueUpdate{
						adapter.ValueUpdate{
							Key:   vt.id + "/volume",
							Value: val,
						},
					},
				})
			}
		}
		vt.Unlock()
	} else {
		return errors.New("Could not get volume from response")
	}

	return nil
}

func (vt *VieraTV) readMute(emitUpdate bool) error {
	b, err := vt.sendCMD("render", "GetMute", "<InstanceID>0</InstanceID><Channel>Master</Channel>")
	if err != nil {
		return err
	}

	vol := currentMuteRegex.FindSubmatch(b)
	if len(vol) > 1 {
		intval, err := strconv.ParseInt(string(vol[1]), 10, 64)
		if err != nil {
			return err
		}

		vt.Lock()
		val := intval != 0
		if val != vt.mute {
			vt.mute = val
			if emitUpdate {
				vt.SendUpdate(adapter.Update{
					ValueContainer: vt,
					Updates: []adapter.ValueUpdate{
						adapter.ValueUpdate{
							Key:   vt.id + "/mute",
							Value: val,
						},
					},
				})
			}
		}
		vt.Unlock()
	} else {
		return errors.New("Could not get volume from response")
	}

	return nil
}

func (vt *VieraTV) sendCMD(typ, action, command string) ([]byte, error) {
	var url string
	var urn string
	switch typ {
	case "control":
		url = "/nrc/control_0"
		urn = "panasonic-com:service:p00NetworkControl:1"
	case "render":
		url = "/dmr/control_0"
		urn = "schemas-upnp-org:service:RenderingControl:1"
	}

	b := &bytes.Buffer{}
	err := cmdTemplate.Execute(b, map[string]string{"urn": urn, "action": action, "command": command})
	if err != nil {
		log.Fatal("[VIERA] cmd template failed:", err)
	}

	client := &http.Client{}
	req, err := http.NewRequest(
		http.MethodPost,
		"http://"+vt.host+url,
		b,
	)
	if err != nil {
		return nil, err
	}

	req.Header.Add("SOAPACTION", "\"urn:"+urn+"#"+action+"\"")

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	if res.StatusCode != 200 {
		return nil, fmt.Errorf("Invalid response from the server (status code %d)", res.StatusCode)
	}

	defer res.Body.Close()
	return ioutil.ReadAll(res.Body)
}

func (vt *VieraTV) Get(id string) (interface{}, error) {
	vt.Lock()
	defer vt.Unlock()

	switch id {
	case "power":
		return vt.power, nil
	case "volume":
		return vt.volume, nil
	case "mute":
		return vt.mute, nil
	}

	return nil, nil
}

func (vt *VieraTV) Set(id string, val interface{}) error {
	switch id {
	case "power":
		if boolval, ok := val.(bool); ok {
			vt.setPower(boolval)
		}
	case "mute":
		if boolval, ok := val.(bool); ok {
			vt.setMute(boolval)
		}
	case "volume":
		switch val := val.(type) {
		case int:
			vt.setVolume(val)
		case int64:
			vt.setVolume(int(val))
		case float64:
			vt.setVolume(int(val))
		}
	}

	return nil
}

func (vt *VieraTV) GetAll() (map[string]interface{}, error) {
	vals := map[string]interface{}{}

	for _, key := range []string{"power", "volume", "mute"} {
		val, err := vt.Get(key)
		if err != nil {
			return nil, err
		}
		vals[key] = val
	}

	return vals, nil
}

func (vt *VieraTV) ID() string {
	return vt.ID()
}

func (vt *VieraTV) UpdateChannel() <-chan adapter.Update {
	return vt.Updater.UpdateChannel()
}
