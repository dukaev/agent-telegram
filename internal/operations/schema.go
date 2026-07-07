package operations

import (
	"reflect"
	"strings"
	"time"
)

var timeType = reflect.TypeOf(time.Time{})

// JSONSchema is a compact JSON Schema representation used in manifests.
type JSONSchema map[string]any

// SchemaFor returns a JSON Schema for a Go struct type.
func SchemaFor(t reflect.Type) JSONSchema {
	return schemaFor(t, false)
}

// InputSchemaFor returns a JSON Schema for request params, including custom
// validation hints such as "peer or username".
func InputSchemaFor(t reflect.Type) JSONSchema {
	return schemaFor(t, true)
}

func schemaFor(t reflect.Type, input bool) JSONSchema {
	if t == nil {
		return objectSchema(nil, nil)
	}
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return schemaForType(t, input).(JSONSchema)
}

func schemaForType(t reflect.Type, input bool) any {
	if t.Kind() == reflect.Ptr {
		return schemaForType(t.Elem(), input)
	}

	switch t.Kind() {
	case reflect.String:
		return JSONSchema{"type": "string"}
	case reflect.Bool:
		return JSONSchema{"type": "boolean"}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32:
		return JSONSchema{"type": "integer"}
	case reflect.Int64:
		return JSONSchema{"type": "integer", "format": "int64"}
	case reflect.Float32, reflect.Float64:
		return JSONSchema{"type": "number"}
	case reflect.Slice:
		return JSONSchema{"type": "array", "items": schemaForType(t.Elem(), input)}
	case reflect.Map:
		return JSONSchema{"type": "object", "additionalProperties": true}
	case reflect.Struct:
		if t == timeType {
			return JSONSchema{"type": "string", "format": "date-time"}
		}
		return structSchema(t, input)
	default:
		return JSONSchema{}
	}
}

func structSchema(t reflect.Type, input bool) JSONSchema {
	properties := map[string]any{}
	required := []string{}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if !field.IsExported() {
			continue
		}

		if field.Anonymous {
			embedded := schemaFor(field.Type, input)
			if props, ok := embedded["properties"].(map[string]any); ok {
				for key, value := range props {
					properties[key] = value
				}
			}
			if req, ok := embedded["required"].([]string); ok {
				required = append(required, req...)
			}
			continue
		}

		name, omit := jsonFieldName(field)
		if name == "" {
			continue
		}
		properties[name] = schemaForType(field.Type, input)
		if isRequired(field) && !omit {
			required = append(required, name)
		}
	}

	schema := objectSchema(properties, required)
	if input {
		if _, hasPeer := properties["peer"]; hasPeer && !contains(required, "peer") {
			if _, hasUsername := properties["username"]; hasUsername && !contains(required, "username") {
				schema["anyOf"] = []map[string][]string{
					{"required": []string{"peer"}},
					{"required": []string{"username"}},
				}
			}
		}
	}
	return schema
}

func objectSchema(properties map[string]any, required []string) JSONSchema {
	if properties == nil {
		properties = map[string]any{}
	}
	schema := JSONSchema{
		"type":                 "object",
		"properties":           properties,
		"additionalProperties": false,
	}
	if len(required) > 0 {
		schema["required"] = required
	}
	return schema
}

func contains(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func jsonFieldName(field reflect.StructField) (name string, omitEmpty bool) {
	tag := field.Tag.Get("json")
	if tag == "-" {
		return "", false
	}
	if tag == "" {
		return field.Name, false
	}
	parts := strings.Split(tag, ",")
	name = parts[0]
	if name == "" {
		name = field.Name
	}
	for _, part := range parts[1:] {
		if part == "omitempty" {
			omitEmpty = true
			break
		}
	}
	return name, omitEmpty
}

func isRequired(field reflect.StructField) bool {
	for _, part := range strings.Split(field.Tag.Get("validate"), ",") {
		if strings.TrimSpace(part) == "required" {
			return true
		}
	}
	return false
}
