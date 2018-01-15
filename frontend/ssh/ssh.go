package ssh

import (
	"encoding/json"
	"io"
	"regexp"
	"strings"

	"github.com/orktes/homeautomation/adapter"
	"github.com/orktes/homeautomation/util"
	"golang.org/x/crypto/ssh/terminal"

	"github.com/gliderlabs/ssh"
	"github.com/orktes/homeautomation/frontend"
	"github.com/orktes/homeautomation/hub"
	"github.com/orktes/homeautomation/registry"
)

var lastExpressionRegex = regexp.MustCompile(`[a-zA-Z0-9]([a-zA-Z0-9\.]*[a-zA-Z0-9])?\.?$`)

type SSHFrontend struct {
	addr     string
	password string
	hub      *hub.Hub
	*ssh.Server
}

func (sshf *SSHFrontend) init() {
	sshf.Server = &ssh.Server{
		Addr:    sshf.addr,
		Handler: ssh.Handler(sshf.handler),
		PasswordHandler: func(ctx ssh.Context, pass string) bool {
			return pass == sshf.password
		},
	}
}

func (sshf *SSHFrontend) handler(s ssh.Session) {

	terminal := terminal.NewTerminal(s, "\033[31m»\033[0m ")

	lastLine := "hub"

	for {
		line, err := terminal.ReadLine()
		if err == io.EOF {
			break
		}

		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		if line == "exit" || line == "logout" {
			break
		}

		if line == "." {
			line = "hub"
		}

		if lastLine != "" && len(line) >= 1 && line[0] == '.' {
			line = lastLine + line
		}

		// TODO only convert member expressions not anything else
		line = util.ConvertDotIDToJavascript(line)

		val, err := sshf.hub.RunScript(line)
		if err != nil {
			terminal.Write([]byte(err.Error()))
			terminal.Write([]byte("\n"))
			continue
		}

		lastLine = "hub"

		if vc, ok := val.(adapter.ValueContainer); ok {
			lastLine = line
			var iterate func(vc adapter.ValueContainer, prefix string)
			iterate = func(vc adapter.ValueContainer, prefix string) {
				all, _ := vc.GetAll()
				for key, item := range all {
					switch item := item.(type) {
					case adapter.ValueContainer:
						iterate(item, prefix+"."+key)
					default:
						terminal.Write([]byte(prefix))
						terminal.Write([]byte("."))
						terminal.Write([]byte(key))
						terminal.Write([]byte(" = "))
						json.NewEncoder(terminal).Encode(item)
					}
				}
			}
			iterate(vc, "")
		} else {
			json.NewEncoder(terminal).Encode(val)
		}
	}
}

func (sshf *SSHFrontend) listen() error {
	return sshf.Server.ListenAndServe()
}

func (sshf *SSHFrontend) Close() error {
	panic("Not implemented")
}

func Create(id string, config map[string]interface{}, hub *hub.Hub) (frontend.Frontend, error) {
	addr := config["addr"].(string)
	password := config["password"].(string)
	f := &SSHFrontend{addr: addr, hub: hub, password: password}
	f.init()
	go f.listen()

	return nil, nil
}

func init() {
	registry.RegisterFrontend("ssh", Create)
}
