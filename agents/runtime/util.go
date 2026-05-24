package runtime

import "encoding/json"

// remarshal re-encodes an interface{} payload (from the event bus) into a typed struct.
// The event bus stores payloads as interface{} which Go decodes as map[string]interface{};
// this round-trip through JSON gives us proper typed structs in each agent.
func remarshal(src interface{}, dst interface{}) error {
	b, err := json.Marshal(src)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, dst)
}
