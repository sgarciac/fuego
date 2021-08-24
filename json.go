package main

import (
	"encoding/base64"
	"encoding/json"
	latlng "google.golang.org/genproto/googleapis/type/latlng"
	"io/ioutil"
	"math"
	"regexp"
	"strings"
	"time"
)

var rfc3339regex, _ = regexp.Compile("^" + rfc3339pattern + "$")

// Return the corresponding firestore primitive value from an extended json
// primitive value, or nil if the value is not one of the extended json forms.
func extendedJsonValueToFirestoreValue(v map[string]interface{}) interface{} {
	if content, ok := v["$boolean"]; ok {
		return content.(bool)
	} else if content, ok := v["$date"]; ok {
		typedContent := content.(string)
		timestamp, err := time.Parse(time.RFC3339Nano, typedContent)
		if err != nil {
			panic("Unexpected error parsing rfc3339")
		}
		return timestamp
	} else if content, ok := v["$numberDouble"]; ok {
		return content.(float64)
	} else if content, ok := v["$numberInt"]; ok {
		typedContent := content.(float64)
		return int64(math.Round(typedContent))
	} else if content, ok := v["$string"]; ok {
		return content.(string)
	} else if content, ok := v["$geopoint"]; ok {
		typedContent := content.(map[string]interface{})
		longitude := typedContent["$longitude"].(float64)
		latitude := typedContent["$latitude"].(float64)
		return &latlng.LatLng{Latitude: latitude, Longitude: longitude}
	} else if content, ok := v["$binary"]; ok {
		typedContent := content.(string)
		data, err := base64.StdEncoding.DecodeString(typedContent)
		if err != nil {
			panic("Unexpected error decoding base64")
		}
		return data
	} else {
		return nil
	}
}

// transformExtendedJsonMapToFirestoreMap traverses a map, replacing extended
// json values by the golang value used by the firestore library to represent
// the given type.
func transformExtendedJsonMapToFirestoreMap(m map[string]interface{}) {
	for k, v := range m {
		switch v := v.(type) {
		case []interface{}:
			transformExtendedJsonArrayToFirestoreArray(v)
		case map[string]interface{}:
			if len(v) == 1 {
				if transformedValue := extendedJsonValueToFirestoreValue(v); transformedValue != nil {
					m[k] = transformedValue
				} else {
					transformExtendedJsonMapToFirestoreMap(v)
				}
			} else {
				transformExtendedJsonMapToFirestoreMap(v)
			}
		}
	}
}

// transformExtendedJsonArrayToFirestoreArray traverses a slice, replacing
// extended json values by the golang value used by the firestore library to
// represent the given type.
func transformExtendedJsonArrayToFirestoreArray(slice []interface{}) {
	for i, v := range slice {
		switch v := v.(type) {
		case []interface{}:
			transformExtendedJsonArrayToFirestoreArray(v)
		case map[string]interface{}:
			if len(v) == 1 {
				if transformedValue := extendedJsonValueToFirestoreValue(v); transformedValue != nil {
					slice[i] = transformedValue
				} else {
					transformExtendedJsonMapToFirestoreMap(v)
				}
			} else {
				transformExtendedJsonMapToFirestoreMap(v)
			}
		}
	}
}

// Returns an extended json primitive value from a firestore primitive value.
func firestoreValueToExtendedJsonValue(v interface{}) interface{} {
	switch v := v.(type) {
	case float64:
		if math.IsNaN(v) {
			return map[string]interface{}{
				"$numberDouble": nil,
			}
		} else {
			return map[string]interface{}{
				"$numberDouble": v,
			}
		}

	case int64:
		return map[string]interface{}{
			"$numberInt": v,
		}

	case bool:
		return map[string]interface{}{
			"$boolean": v,
		}

	case time.Time:
		return map[string]interface{}{
			"$date": v.Format(time.RFC3339Nano),
		}

	case string:
		return map[string]interface{}{
			"$string": v,
		}

	case *latlng.LatLng:
		return map[string]interface{}{
			"$geopoint": map[string]interface{}{
				"$latitude":  v.Latitude,
				"$longitude": v.Longitude,
			},
		}
	case []byte:
		return map[string]interface{}{
			"$binary": base64.StdEncoding.EncodeToString(v),
		}
	default:
		return v
	}

}

// transformFirestoreMapToExtendedJsonMap transforms a firestore map to an extended json map.
func transformFirestoreMapToExtendedJsonMap(m map[string]interface{}) {
	for k, v := range m {
		switch v := v.(type) {
		case []interface{}:
			transformFirestoreArrayToExtendedJsonArray(v)
		case map[string]interface{}:
			transformFirestoreMapToExtendedJsonMap(v)
		default:
			m[k] = firestoreValueToExtendedJsonValue(v)
		}
	}
}

func transformFirestoreArrayToExtendedJsonArray(slice []interface{}) {
	for i, v := range slice {
		switch v := v.(type) {
		case []interface{}:
			transformFirestoreArrayToExtendedJsonArray(v)
		case map[string]interface{}:
			transformFirestoreMapToExtendedJsonMap(v)
		default:
			slice[i] = firestoreValueToExtendedJsonValue(v)
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

func marshallData(object map[string]interface{}, extended bool) (string, error) {
	if extended {
		transformFirestoreMapToExtendedJsonMap(object)
	} else {
		unNaNMap(object)
	}

	buffer, err := json.MarshalIndent(object, "", "    ")
	if err != nil {
		return "", err
	}
	return string(buffer), nil
}
