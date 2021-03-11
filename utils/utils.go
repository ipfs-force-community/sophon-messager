package utils

import "encoding/json"

func ObjectToString(i interface{}) string {
	res, _ := json.MarshalIndent(i, "", " ")
	return string(res)
}
