package main

import (
	"encoding/json"
	"io/ioutil"
	"math"
	"regexp"
	"strings"
	"time"
)

var rfc3339regex, _ = regexp.Compile("^" + rfc3339pattern + "$")

// Replaces all string values in an array that are RFC3339 timestamps
// by its corresponding time.Time object, recursively.
func timestampifySlice(slice []interface{}) {
	for i, v := range slice {
		switch v := v.(type) {
		case []interface{}:
			timestampifySlice(v)
		case map[string]interface{}:
			timestampifyMap(v)
		case string:
			if rfc3339regex.MatchString(v) {
				timestamp, err := time.Parse(time.RFC3339Nano, v)
				if err != nil {
					panic("Unexpected error parsing rfc3339")
				}
				slice[i] = timestamp
			}
		}
	}
}

// Replaces all string values in an object that are RFC3339 timestamps
// by its corresponding time.Time object, recursively.
func timestampifyMap(m map[string]interface{}) {
	for k, v := range m {
		switch v := v.(type) {
		case []interface{}:
			timestampifySlice(v)
		case map[string]interface{}:
			timestampifyMap(v)
		case string:
			if rfc3339regex.MatchString(v) {
				// hoping our regex prevents wrong parsing
				timestamp, err := time.Parse(time.RFC3339, v)
				if err != nil {
					panic("Unexpected error parsing rfc3339")
				}
				m[k] = timestamp
			}
		}
	}
}

// Replaces all NaN values in an array by null, recursively.
func unNaNSlice(slice []interface{}) {
	for i, v := range slice {
		switch v := v.(type) {
		case []interface{}:
			timestampifySlice(v)
		case map[string]interface{}:
			timestampifyMap(v)
		case float64:
			if math.IsNaN(v) {
				slice[i] = nil
			}
		}
	}
}

// Replaces all NaN values in an object by null, recursively.
func unNaNMap(m map[string]interface{}) {
	for k, v := range m {
		switch v := v.(type) {
		case []interface{}:
			timestampifySlice(v)
		case map[string]interface{}:
			timestampifyMap(v)
		case float64:
			if math.IsNaN(v) {
				m[k] = nil
			}
		}
	}
}

// unmarshall data
func unmarshallData(data string) (map[string]interface{}, error) {
	trimmed := strings.TrimSpace(data)
	var buffer []byte
	if strings.HasPrefix(trimmed, "{") {
		buffer = []byte(trimmed)
	} else {
		var err error
		buffer, err = ioutil.ReadFile(data)
		if err != nil {
			return nil, err
		}
	}
	var object map[string]interface{}
	err := json.Unmarshal(buffer, &object)
	if err != nil {
		return nil, err
	}
	return object, nil
}

func marshallData(object map[string]interface{}) (string, error) {
	unNaNMap(object)
	buffer, err := json.MarshalIndent(object, "", "    ")
	if err != nil {
		return "", err
	}
	return string(buffer), nil
}
