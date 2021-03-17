package utils

import (
	"encoding/json"
	"reflect"
)

func Unmarshal(in interface{}, raw []byte, callback func() error) error {
	if err := json.Unmarshal(raw, &in); err != nil {
		return err
	}
	return callback()
}

func JsonStringAutoDecode(m interface{}) func(rf reflect.Kind, rt reflect.Kind, data interface{}) (interface{}, error) {
	return func(rf reflect.Kind, rt reflect.Kind, data interface{}) (interface{}, error) {
		if rf != reflect.String || rt == reflect.String {
			return data, nil
		}

		raw := data.(string)
		if raw != "" && (raw[0:1] == "{" || raw[0:1] == "[") {
			err := json.Unmarshal([]byte(raw), &m)
			return m, err
		}

		return data, nil
	}
}
