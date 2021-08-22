package main

import (
	"encoding/json"
	latlng "google.golang.org/genproto/googleapis/type/latlng"
	"io/ioutil"
	"math"
	"regexp"
	"strings"
	"time"
)

var rfc3339regex, _ = regexp.Compile("^" + rfc3339pattern + "$")

// prepareTypesMap traverses a map, replacing extended json values
// by the golang value used by the firestore library to represent the given type.
func prepareTypesMap(m map[string]interface{}) {
	for k, v := range m {
		switch v := v.(type) {
		case []interface{}:
			prepareTypesSlice(v)
		case map[string]interface{}:
			if len(v) == 1 {
				if content, ok := v["$boolean"]; ok {
					typedContent := content.(bool)
					m[k] = typedContent
				} else if content, ok := v["$date"]; ok {
					typedContent := content.(string)
					timestamp, err := time.Parse(time.RFC3339Nano, typedContent)
					if err != nil {
						panic("Unexpected error parsing rfc3339")
					}
					m[k] = timestamp
				} else if content, ok := v["$numberDouble"]; ok {
					typedContent := content.(float64)
					m[k] = typedContent
				} else if content, ok := v["$numberInt"]; ok {
					typedContent := content.(float64)
					m[k] = int64(math.Round(typedContent))
				} else if content, ok := v["$string"]; ok {
					typedContent := content.(string)
					m[k] = typedContent
				} else if content, ok := v["$geopoint"]; ok {
					typedContent := content.(map[string]interface{})
					longitude := typedContent["$longitude"].(float64)
					latitude := typedContent["$latitude"].(float64)
					m[k] = &latlng.LatLng{Latitude: latitude, Longitude: longitude}
				}

			} else {
				prepareTypesMap(v)
			}
		}
	}
}

// prepareTypesMap traverses a slice, replacing extended json values
// by the golang value used by the firestore library to represent the given type.
func prepareTypesSlice(slice []interface{}) {
	for i, v := range slice {
		switch v := v.(type) {
		case []interface{}:
			prepareTypesSlice(v)
		case map[string]interface{}:
			if len(v) == 1 {
				if content, ok := v["$boolean"]; ok {
					typedContent := content.(bool)
					slice[i] = typedContent
				} else if content, ok := v["$date"]; ok {
					typedContent := content.(string)
					timestamp, err := time.Parse(time.RFC3339Nano, typedContent)
					if err != nil {
						panic("Unexpected error parsing rfc3339")
					}
					slice[i] = timestamp
				} else if content, ok := v["$numberDouble"]; ok {
					typedContent := content.(float64)
					slice[i] = typedContent
				} else if content, ok := v["$numberInt"]; ok {
					typedContent := content.(float64)
					slice[i] = int64(math.Round(typedContent))
				} else if content, ok := v["$string"]; ok {
					typedContent := content.(string)
					slice[i] = typedContent
				} else if content, ok := v["$geopoint"]; ok {
					typedContent := content.(map[string]interface{})
					longitude := typedContent["$longitude"].(float64)
					latitude := typedContent["$latitude"].(float64)
					slice[i] = &latlng.LatLng{Latitude: latitude, Longitude: longitude}
				}

			} else {
				prepareTypesMap(v)
			}
		}
	}
}

// Replaces all NaN values in an array by null, recursively.
func unNaNSlice(slice []interface{}) {
	for i, v := range slice {
		switch v := v.(type) {
		case []interface{}:
			unNaNSlice(v)
		case map[string]interface{}:
			unNaNMap(v)
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
			unNaNSlice(v)
		case map[string]interface{}:
			unNaNMap(v)
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
