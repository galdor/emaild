package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"

	"github.com/galdor/emaild/pkg/smtp"
	"github.com/galdor/go-ejson"
	"github.com/galdor/go-log"
	"gopkg.in/yaml.v3"
)

type ServerCfg struct {
	BuildId string `json:"-"`

	Logger      *log.LoggerCfg            `json:"logger"`
	SMTPServers map[string]smtp.ServerCfg `json:"smtp_servers"`
}

func (cfg *ServerCfg) Load(filePath string) error {
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("cannot read %q: %w", filePath, err)
	}

	yamlDecoder := yaml.NewDecoder(bytes.NewReader(data))

	var yamlValue any
	if err := yamlDecoder.Decode(&yamlValue); err != nil && err != io.EOF {
		return fmt.Errorf("cannot decode yaml data: %w", err)
	}

	jsonValue, err := YAMLValueToJSONValue(yamlValue)
	if err != nil {
		return fmt.Errorf("invalid yaml data: %w", err)
	}

	jsonData, err := json.Marshal(jsonValue)
	if err != nil {
		return fmt.Errorf("cannot generate json data: %w", err)
	}

	if err := ejson.Unmarshal(jsonData, cfg); err != nil {
		return fmt.Errorf("cannot decode json data: %w", err)
	}

	return nil
}

func YAMLValueToJSONValue(yamlValue any) (any, error) {
	// For some reason, go-yaml will return objects as map[string]any
	// if all keys are strings, and as map[any]any if not. So
	// we have to handle both.

	var jsonValue any

	switch v := yamlValue.(type) {
	case []any:
		array := make([]any, len(v))

		for i, yamlElement := range v {
			jsonElement, err := YAMLValueToJSONValue(yamlElement)
			if err != nil {
				return nil, err
			}

			array[i] = jsonElement
		}

		jsonValue = array

	case map[any]any:
		object := make(map[string]any)

		for key, yamlEntry := range v {
			keyString, ok := key.(string)
			if !ok {
				return nil,
					fmt.Errorf("object key \"%v\" is not a string", key)
			}

			jsonEntry, err := YAMLValueToJSONValue(yamlEntry)
			if err != nil {
				return nil, err
			}

			object[keyString] = jsonEntry
		}

		jsonValue = object

	case map[string]any:
		object := make(map[string]any)

		for key, yamlEntry := range v {
			jsonEntry, err := YAMLValueToJSONValue(yamlEntry)
			if err != nil {
				return nil, err
			}

			object[key] = jsonEntry
		}

		jsonValue = object

	default:
		jsonValue = yamlValue
	}

	return jsonValue, nil
}
