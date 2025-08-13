package parseform

import (
	"fmt"
	"net/url"
	"reflect"
	"strconv"
	"strings"
)

// Parser represents a form-urlencoded data parser
type Parser struct{}

// NewParser creates a new parser instance
func NewParser() *Parser {
	return &Parser{}
}

// ParseForm parses form-urlencoded data into a struct
func (p *Parser) ParseForm(formData string, target interface{}) error {
	// Parse the form data
	values, err := url.ParseQuery(formData)
	if err != nil {
		return fmt.Errorf("failed to parse form data: %w", err)
	}

	// Parse into target struct
	return p.parseIntoStruct(values, target)
}

// ParseFormBytes parses form-urlencoded data from bytes into a struct
func (p *Parser) ParseFormBytes(data []byte, target interface{}) error {
	return p.ParseForm(string(data), target)
}

// parseIntoStruct parses url.Values data into a struct
func (p *Parser) parseIntoStruct(values url.Values, target interface{}) error {
	targetValue := reflect.ValueOf(target)
	if targetValue.Kind() != reflect.Ptr || targetValue.IsNil() {
		return fmt.Errorf("target must be a non-nil pointer to struct")
	}

	targetElem := targetValue.Elem()
	if targetElem.Kind() != reflect.Struct {
		return fmt.Errorf("target must be a pointer to struct")
	}

	return p.parseStruct(values, targetElem)
}

// parseStruct recursively parses data into a struct
func (p *Parser) parseStruct(values url.Values, structValue reflect.Value) error {
	structType := structValue.Type()

	for i := 0; i < structValue.NumField(); i++ {
		field := structValue.Field(i)
		fieldType := structType.Field(i)

		// Get the form tag or use field name
		formTag := fieldType.Tag.Get("form")
		fieldName := fieldType.Name
		if formTag != "" {
			fieldName = formTag
		}

		// Try to find matching data for this field
		fieldData := p.findFieldData(values, fieldName)
		if fieldData == nil {
			continue
		}

		// Parse the field value
		if err := p.parseFieldValue(field, fieldData, fieldName); err != nil {
			return fmt.Errorf("failed to parse field %s: %w", fieldName, err)
		}
	}

	return nil
}

// findFieldData finds data that matches a field name (including nested notation)
func (p *Parser) findFieldData(values url.Values, fieldName string) map[string]string {
	result := make(map[string]string)

	// Look for exact matches and nested matches
	for key, valueSlice := range values {
		if len(valueSlice) == 0 {
			continue
		}

		if key == fieldName {
			result[key] = valueSlice[0]
		} else if strings.HasPrefix(key, fieldName+"[") {
			// Extract nested part - keep the full nested key with brackets
			nestedKey := key[len(fieldName)+1:] // Remove fieldName[ but keep the rest
			result[nestedKey] = valueSlice[0]
		}
	}

	if len(result) == 0 {
		return nil
	}

	return result
}

// parseFieldValue parses a single field value
func (p *Parser) parseFieldValue(field reflect.Value, fieldData map[string]string, fieldName string) error {
	// Handle different field types
	switch field.Kind() {
	case reflect.String:
		for _, value := range fieldData {
			field.SetString(value)
			return nil
		}

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		for _, value := range fieldData {
			if intVal, err := strconv.ParseInt(value, 10, 64); err == nil {
				field.SetInt(intVal)
				return nil
			}
		}

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		for _, value := range fieldData {
			if uintVal, err := strconv.ParseUint(value, 10, 64); err == nil {
				field.SetUint(uintVal)
				return nil
			}
		}

	case reflect.Float32, reflect.Float64:
		for _, value := range fieldData {
			if floatVal, err := strconv.ParseFloat(value, 64); err == nil {
				field.SetFloat(floatVal)
				return nil
			}
		}

	case reflect.Bool:
		for _, value := range fieldData {
			if boolVal, err := strconv.ParseBool(value); err == nil {
				field.SetBool(boolVal)
				return nil
			}
		}

	case reflect.Struct:
		// Handle nested structs
		if field.CanSet() {
			// Create a new instance of the struct type
			newStruct := reflect.New(field.Type()).Elem()
			if err := p.parseStructFromMap(fieldData, newStruct); err == nil {
				field.Set(newStruct)
				return nil
			}
		}

	case reflect.Slice:
		// Handle slices
		return p.parseSlice(field, fieldData)

	case reflect.Map:
		// Handle maps
		return p.parseMap(field, fieldData, fieldName)
	}

	return nil
}

// parseStructFromMap parses a struct from a map of field data
func (p *Parser) parseStructFromMap(fieldData map[string]string, structValue reflect.Value) error {
	structType := structValue.Type()

	for i := 0; i < structValue.NumField(); i++ {
		field := structValue.Field(i)
		fieldType := structType.Field(i)

		// Get the form tag or use field name
		formTag := fieldType.Tag.Get("form")
		fieldName := fieldType.Name
		if formTag != "" {
			fieldName = formTag
		}

		// Try to find matching data for this field
		if value, exists := fieldData[fieldName]; exists {
			if err := p.setValue(field, value); err != nil {
				continue
			}
		}
	}

	return nil
}

// parseSlice parses slice fields
func (p *Parser) parseSlice(field reflect.Value, fieldData map[string]string) error {
	// Group data by index
	indexedData := make(map[int]map[string]string)

	for key, value := range fieldData {
		if strings.Contains(key, "[") && strings.Contains(key, "]") {
			// Extract index from key like "field[0][subfield]"
			parts := strings.Split(key, "[")
			if len(parts) >= 2 {
				indexStr := strings.TrimSuffix(parts[1], "]")
				if index, err := strconv.Atoi(indexStr); err == nil {
					if indexedData[index] == nil {
						indexedData[index] = make(map[string]string)
					}

					// Reconstruct the nested key
					nestedKey := strings.Join(parts[2:], "[")
					if nestedKey != "" {
						indexedData[index][nestedKey] = value
					} else {
						indexedData[index]["value"] = value
					}
				}
			}
		}
	}

	// Create slice with appropriate length
	if len(indexedData) > 0 {
		sliceType := field.Type()
		elemType := sliceType.Elem()

		// Find the maximum index to determine slice length
		maxIndex := 0
		for index := range indexedData {
			if index > maxIndex {
				maxIndex = index
			}
		}

		// Create slice with maxIndex + 1 elements
		slice := reflect.MakeSlice(sliceType, maxIndex+1, maxIndex+1)

		// Parse each element
		for index, data := range indexedData {
			if index < slice.Len() {
				elem := slice.Index(index)

				switch elemType.Kind() {
				case reflect.Struct:
					newElem := reflect.New(elemType).Elem()
					if err := p.parseStructFromMap(data, newElem); err == nil {
						elem.Set(newElem)
					}
				case reflect.String:
					if value, exists := data["value"]; exists {
						elem.SetString(value)
					}
				case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
					if value, exists := data["value"]; exists {
						if intVal, err := strconv.ParseInt(value, 10, 64); err == nil {
							elem.SetInt(intVal)
						}
					}
				}
			}
		}

		field.Set(slice)
	}

	return nil
}

// parseMap parses map fields
func (p *Parser) parseMap(field reflect.Value, fieldData map[string]string, fieldName string) error {
	// Group data by map key
	mapData := make(map[string]string)

	for key, value := range fieldData {
		if strings.HasPrefix(key, fieldName+"[") && strings.HasSuffix(key, "]") {
			// Extract map key from "field[key]"
			mapKey := key[len(fieldName)+1 : len(key)-1]
			mapData[mapKey] = value
		}
	}

	// Create map and populate it
	if len(mapData) > 0 {
		mapType := field.Type()
		keyType := mapType.Key()
		elemType := mapType.Elem()

		newMap := reflect.MakeMap(mapType)

		for keyStr, valueStr := range mapData {
			// Parse key
			keyValue := reflect.New(keyType).Elem()
			if err := p.setValue(keyValue, keyStr); err != nil {
				continue
			}

			// Parse value
			elemValue := reflect.New(elemType).Elem()
			if err := p.setValue(elemValue, valueStr); err != nil {
				continue
			}

			newMap.SetMapIndex(keyValue, elemValue)
		}

		field.Set(newMap)
	}

	return nil
}

// setValue sets a value to a reflect.Value based on its type
func (p *Parser) setValue(field reflect.Value, value string) error {
	switch field.Kind() {
	case reflect.String:
		field.SetString(value)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if intVal, err := strconv.ParseInt(value, 10, 64); err == nil {
			field.SetInt(intVal)
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if uintVal, err := strconv.ParseUint(value, 10, 64); err == nil {
			field.SetUint(uintVal)
		}
	case reflect.Float32, reflect.Float64:
		if floatVal, err := strconv.ParseFloat(value, 64); err == nil {
			field.SetFloat(floatVal)
		}
	case reflect.Bool:
		if boolVal, err := strconv.ParseBool(value); err == nil {
			field.SetBool(boolVal)
		}
	}
	return nil
}

// Utility functions for common parsing needs

// ParseTimestamp parses Unix timestamp to time.Time
func ParseTimestamp(timestamp string) (int64, error) {
	if timestamp == "" {
		return 0, fmt.Errorf("empty timestamp")
	}

	ts, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid timestamp format: %w", err)
	}

	return ts, nil
}

// ParseInt parses string to int with error handling
func ParseInt(value string) (int, error) {
	if value == "" {
		return 0, fmt.Errorf("empty value")
	}

	return strconv.Atoi(value)
}

// ParseFloat parses string to float64 with error handling
func ParseFloat(value string) (float64, error) {
	if value == "" {
		return 0, fmt.Errorf("empty value")
	}

	return strconv.ParseFloat(value, 64)
}
