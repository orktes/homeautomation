package structs

import "encoding/json"

type Request struct {
	ID      string
	Topic   string
	Payload json.RawMessage
}
