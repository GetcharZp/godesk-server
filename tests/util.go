package tests

import "encoding/json"

func prettyPrint(v any) string {
	b, _ := json.MarshalIndent(v, "", "	")
	return string(b)
}
