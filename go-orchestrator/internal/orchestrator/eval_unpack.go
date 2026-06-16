package orchestrator

import "encoding/json"

// unwrapExtendScriptEnvelope strips the _ok()/panel wrapper {"success":true,"data":...}
// when present, so unmarshaling into protobuf-shaped structs succeeds.
func unwrapExtendScriptEnvelope(result string) []byte {
	var top struct {
		Success bool            `json:"success"`
		Data    json.RawMessage `json:"data"`
	}
	if json.Unmarshal([]byte(result), &top) != nil || !top.Success || len(top.Data) == 0 {
		return []byte(result)
	}
	return top.Data
}
