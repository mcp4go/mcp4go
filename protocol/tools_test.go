package protocol

import (
	"encoding/json"
	"testing"

	"github.com/mcp4go/mcp4go/protocol/jsonschema"
)

func TestMarshalTools(t *testing.T) {
	bs, _ := json.Marshal(Tool{})
	if string(bs) != "{\"name\":\"\"}" {
		t.Errorf("unexpected json: %s", string(bs))
	}
	bs, _ = json.Marshal(Tool{
		Name:        "",
		Description: "",
		InputSchema: &jsonschema.Definition{
			Type: jsonschema.Object,
			Properties: map[string]jsonschema.Definition{
				"foo": {Type: jsonschema.String},
			},
		},
	})
	if string(bs) != "{\"name\":\"\",\"inputSchema\":{\"type\":\"object\",\"properties\":{\"foo\":{\"type\":\"string\"}}}}" {
		t.Errorf("unexpected json: %s", string(bs))
	}
	bs, _ = json.Marshal(Tool{
		Name:        "",
		Description: "",
		InputSchema: &jsonschema.Definition{
			Type: jsonschema.Object,
		},
	})
	if string(bs) != "{\"name\":\"\",\"inputSchema\":{\"type\":\"object\"}}" {
		t.Errorf("unexpected json: %s", string(bs))
	}
}
