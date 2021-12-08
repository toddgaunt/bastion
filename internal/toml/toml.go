package toml

import (
	"bytes"
	"errors"
	"fmt"
	"reflect"
	"strings"
)

func Marshal(v interface{}) ([]byte, error) {

	value := reflect.ValueOf(v)
	valueType := reflect.TypeOf(v)

	if valueType.Kind() != reflect.Struct {
		return nil, errors.New("must be of kind Struct")
	}

	var buf bytes.Buffer

	for i := 0; i < valueType.NumField(); i += 1 {
		marshalField(&buf, "", valueType.Field(i), value.Field(i))
	}

	return buf.Bytes(), nil
}

func marshalField(buf *bytes.Buffer, parentName string, field reflect.StructField, value reflect.Value) error {
	valueType := value.Type()
	fieldName := strings.ToLower(field.Name)

	switch valueType.Kind() {
	case reflect.Ptr:
		// Unpack pointers to their underlying values if not nil, since a pointer
		// here represents optional values to marshal.
		if !value.IsNil() {
			marshalField(buf, parentName, field, value.Elem())
		}
	case reflect.Int:
		fmt.Fprintf(buf, "%s = %d\n", fieldName, value.Int())
	case reflect.String:
		fmt.Fprintf(buf, "%s = %s\n", fieldName, value.String())
	case reflect.Struct:
		var tableName = fieldName
		if parentName != "" {
			tableName = fmt.Sprintf("%s.%s", parentName, fieldName)
		}
		fmt.Fprintf(buf, "\n[%s]\n", tableName)
		for i := 0; i < valueType.NumField(); i += 1 {
			marshalField(buf, tableName, valueType.Field(i), value.Field(i))
		}
	default:
		return fmt.Errorf("toml: marshal unsupported type %v", valueType)
	}

	return nil
}

func Unmarshal(data []byte, v interface{}) error {
	return nil
}