package hue

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"text/template"

	"github.com/gin-gonic/gin"
	"github.com/orktes/homeautomation/frontend"
	"github.com/orktes/homeautomation/hub"
	"github.com/orktes/homeautomation/registry"
	"github.com/orktes/homeautomation/util"
)

var setupTemplate = template.Must(template.New("setup").Parse(`<?xml version="1.0"?>
<root xmlns="urn:schemas-upnp-org:device-1-0">
	<specVersion>
	<major>1</major>
	<minor>0</minor>
	</specVersion>
	<URLBase>http://{{.IP}}:{{.Port}}/</URLBase>
	<device>
	<deviceType>urn:schemas-upnp-org:device:Basic:1</deviceType>
	<friendlyName>{{.Name}}</friendlyName>
	<manufacturer>Royal Philips Electronics</manufacturer>
	<modelName>Philips hue bridge 2012</modelName>
	<modelNumber>929000226503</modelNumber>
	<UDN>uuid:{{.UUID}}</UDN>
	</device>
</root>`))

type Hue struct {
	Port   int
	UUID   string
	IP     string
	Name   string
	router *gin.Engine
	server *http.Server
	hub    *hub.Hub
}

func (hue *Hue) init() error {
	if err := hue.initUPNP(); err != nil {
		return err
	} else if hue.initServer(); err != nil {
		return err
	}

	return nil
}

func (hue *Hue) initUPNP() error {
	// TODO make upnp multicast addr configurable
	createUPNPResponder(fmt.Sprint("http://%s:%d/upnp/setup.xml", hue.IP, hue.Port), hue.UUID, "239.255.255.250:1900")
	return nil
}

func (hue *Hue) initServer() error {
	router := gin.Default()
	router.GET("/upnp/setup.xml", hue.serveSetupXML)
	router.GET("/api/:userId", hue.getLights)
	router.GET("/api/:userId/lights", hue.getLights)
	router.PUT("/api/:userId/lights/:lightId/state", hue.setLightState)
	router.GET("/api/:userId/lights/:lightId", hue.getLight)
	hue.router = router
	hue.server = &http.Server{
		Addr:    fmt.Sprintf(":%d", hue.Port),
		Handler: router,
	}

	go func() {
		if err := hue.server.ListenAndServe(); err != nil {
			log.Printf("listen: %s\n", err)
		}
	}()

	return nil
}

func (hue *Hue) getLights(c *gin.Context) {
	lightMap := LightList{}

	lights := hue.hub.GetLights()

	for _, light := range lights {
		huelight, err := hue.convertLightToHueFormat(light)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		lightMap[light.ID()] = huelight
	}

	c.JSON(200, lights)
}

func (hue *Hue) getLight(c *gin.Context) {
	id := c.Param("lightId")
	lights := hue.hub.GetLights()

	for _, light := range lights {
		if light.ID() == id {
			huelight, err := hue.convertLightToHueFormat(light)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			c.JSON(200, huelight)
		}
	}

	c.JSON(http.StatusNotFound, gin.H{"error": "light was not found"})
}

func (hue *Hue) setLightState(c *gin.Context) {
	id := c.Param("lightId")
	lights := hue.hub.GetLights()

	for _, light := range lights {
		if light.ID() == id {
			state := &LightStateChange{}
			if err := c.BindJSON(state); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			}

			// Check different states and act accordinly
			if state.On != nil {
				light.SetOn(*state.On)
				// TODO process errors
			}

			if state.Bri != nil {
				light.SetBrightness(*state.Bri)
				// TODO process errors
			}

			resp := &LightStateChangeResponse{}
			// TODO set correct state
			c.JSON(200, resp)
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{"error": "light was not found"})
}

func (hue *Hue) serveSetupXML(c *gin.Context) {
	b := &bytes.Buffer{}
	err := setupTemplate.Execute(b, hue)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.String(http.StatusOK, b.String())
}

func (hue *Hue) convertLightToHueFormat(l *hub.Light) (Light, error) {
	return Light{}, nil
}

func (hue *Hue) Close() error {
	return hue.server.Close()
}

func Create(id string, config map[string]interface{}, hub *hub.Hub) (frontend.Frontend, error) {
	port := config["port"].(int)
	uuid := config["uuid"].(string)

	ip := util.GetIPAddress()

	if ipOverride, ok := config["ip"]; ok {
		ip = ipOverride.(string)
	}

	h := &Hue{
		Port: port,
		UUID: uuid,
		IP:   ip,
		Name: id,
		hub:  hub,
	}

	return h, h.init()
}

func init() {
	registry.RegisterFrontend("hue", Create)
}
