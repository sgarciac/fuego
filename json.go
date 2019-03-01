package main

import (
	"encoding/json"
	"io/ioutil"
	"regexp"
	"strings"
	"time"
)

var rfc3339regex, _ = regexp.Compile("^" + rfc3339pattern + "$")

// Replaces all string values in an object that are RFC3339 timestamps
// by its corresponding time.Time object
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
			} else {
				slice[i] = v
			}
			slice[i] = strings.ToUpper(v)
		}
	}
}

// Replaces all string values in an object that are RFC3339 timestamps
// by its corresponding time.Time object
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
			} else {
				m[k] = v
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

func marshallData(object interface{}) (string, error) {
	buffer, err := json.MarshalIndent(object, "", "    ")
	if err != nil {
		return "", err
	}
	return string(buffer), nil
}
