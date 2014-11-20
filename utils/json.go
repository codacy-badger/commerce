package utils

import (
	"encoding/json"
	"errors"
)

// EncodeToJSONString encodes inputData to JSON string if it's possible
func EncodeToJSONString(inputData interface{}) string {
	result, _ := json.Marshal(inputData)
	return string(result)
}

// DecodeJSONToArray decodes json string to []interface{} if it's possible
func DecodeJSONToArray(jsonData interface{}) ([]interface{}, error) {
	var result []interface{}

	var err error
	switch value := jsonData.(type) {
	case string:
		err = json.Unmarshal([]byte(value), &result)
	case []byte:
		err = json.Unmarshal(value, &result)
	default:
		err = errors.New("unsupported json data")
	}

	return result, err
}

// DecodeJSONToStringKeyMap decodes json string to map[string]interface{} if it's possible
func DecodeJSONToStringKeyMap(jsonData interface{}) (map[string]interface{}, error) {

	result := make(map[string]interface{})

	var err error

	switch value := jsonData.(type) {
	case string:
		err = json.Unmarshal([]byte(value), &result)
	case []byte:
		err = json.Unmarshal(value, &result)
	default:
		err = errors.New("unsupported json data")
	}

	return result, err
}
