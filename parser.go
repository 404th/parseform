package parseform

import (
	"encoding/json"
	"fmt"
	"net/url"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

// Parser represents a form-urlencoded data parser
type Parser struct{}

// keyGroup represents a group of related form keys
type keyGroup struct {
	baseKey   string
	value     interface{} // Change from string to interface{}
	isSimple  bool
	isArray   bool
	isObject  bool
	children  map[string]*keyGroup
	arrayData map[int]*keyGroup
}

// parsedKey represents a parsed form key
type parsedKey struct {
	baseKey    string
	isArray    bool
	isNested   bool
	arrayIndex int
	path       []string
}

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

// FormToJSON converts form-urlencoded data to JSON dynamically
func (p *Parser) FormToJSON(formData string) ([]byte, error) {
	// Parse the form data
	values, err := url.ParseQuery(formData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse form data: %w", err)
	}

	// Convert to dynamic JSON structure
	result := p.parseFormFlexibly(values)

	// Convert to JSON
	jsonData, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal to JSON: %w", err)
	}

	return jsonData, nil
}

// FormToJSONBytes converts form-urlencoded data from bytes to JSON
func (p *Parser) FormToJSONBytes(data []byte) ([]byte, error) {
	return p.FormToJSON(string(data))
}

// FormToMap converts form-urlencoded data to a map[string]interface{} dynamically
func (p *Parser) FormToMap(formData string) (map[string]interface{}, error) {
	// Parse the form data
	values, err := url.ParseQuery(formData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse form data: %w", err)
	}

	// Convert to dynamic map structure
	result := p.parseFormFlexibly(values)
	return result, nil
}

// FormToMapBytes converts form-urlencoded data from bytes to a map
func (p *Parser) FormToMapBytes(data []byte) (map[string]interface{}, error) {
	return p.FormToMap(string(data))
}

// parseFormFlexibly parses any form data structure dynamically
func (p *Parser) parseFormFlexibly(values url.Values) map[string]interface{} {
	result := make(map[string]interface{})

	// Group all keys by their base structure
	keyGroups := p.groupKeysByStructure(values)

	// Process each group
	for baseKey, group := range keyGroups {
		if group.isSimple {
			// Simple key-value pair
			result[baseKey] = group.value
		} else if group.isArray {
			// Array structure
			result[baseKey] = p.buildArrayFromGroup(group)
		} else {
			// Nested object structure
			result[baseKey] = p.buildObjectFromGroup(group)
		}
	}

	return result
}

// groupKeysByStructure groups form keys by their structure
func (p *Parser) groupKeysByStructure(values url.Values) map[string]*keyGroup {
	groups := make(map[string]*keyGroup)

	for key, valueSlice := range values {
		if len(valueSlice) == 0 {
			continue
		}

		value := valueSlice[0]

		// Parse the key structure
		parsed := p.parseKeyStructure(key)

		// Get or create the base group
		if groups[parsed.baseKey] == nil {
			groups[parsed.baseKey] = &keyGroup{
				baseKey:   parsed.baseKey,
				children:  make(map[string]*keyGroup),
				arrayData: make(map[int]*keyGroup),
			}
		}

		group := groups[parsed.baseKey]

		if parsed.isArray {
			group.isArray = true
			p.addToArrayGroup(group, parsed, value)
		} else if parsed.isNested {
			group.isObject = true
			p.addToObjectGroup(group, parsed, value)
		} else {
			group.isSimple = true
			group.value = value
		}
	}

	return groups
}

// parseKeyStructure parses any key format dynamically
func (p *Parser) parseKeyStructure(key string) *parsedKey {
	result := &parsedKey{
		path: make([]string, 0),
	}

	// Handle simple keys
	if !strings.Contains(key, "[") && !strings.Contains(key, "]") {
		result.baseKey = key
		return result
	}

	// Extract base key (everything before first [)
	openBracket := strings.Index(key, "[")
	result.baseKey = key[:openBracket]

	// Parse the rest using regex to find all bracket groups
	re := regexp.MustCompile(`\[([^\]]+)\]`)
	matches := re.FindAllStringSubmatch(key[openBracket:], -1)

	if len(matches) == 0 {
		return result
	}

	// Check if first bracket contains a number (array index)
	if firstMatch := matches[0][1]; p.isNumeric(firstMatch) {
		result.isArray = true
		result.arrayIndex, _ = strconv.Atoi(firstMatch)

		// Add remaining path elements
		for i := 1; i < len(matches); i++ {
			result.path = append(result.path, matches[i][1])
		}
	} else {
		result.isNested = true
		// Add all path elements
		for _, match := range matches {
			result.path = append(result.path, match[1])
		}
	}

	return result
}

// addToArrayGroup adds data to an array group
func (p *Parser) addToArrayGroup(group *keyGroup, parsed *parsedKey, value string) {
	if group.arrayData[parsed.arrayIndex] == nil {
		group.arrayData[parsed.arrayIndex] = &keyGroup{
			baseKey:  fmt.Sprintf("%d", parsed.arrayIndex),
			children: make(map[string]*keyGroup),
		}
	}

	arrayItem := group.arrayData[parsed.arrayIndex]

	if len(parsed.path) == 0 {
		// Direct value at this index
		arrayItem.value = value
		arrayItem.isSimple = true
	} else {
		// Nested structure at this index
		p.addNestedToGroup(arrayItem, parsed.path, value)
	}
}

// addToObjectGroup adds data to an object group
func (p *Parser) addToObjectGroup(group *keyGroup, parsed *parsedKey, value string) {
	if len(parsed.path) == 0 {
		// Direct nested value
		group.value = value
		group.isSimple = true
	} else {
		// Nested structure
		p.addNestedToGroup(group, parsed.path, value)
	}
}

// addNestedToGroup adds nested data to a group
func (p *Parser) addNestedToGroup(group *keyGroup, path []string, value string) {
	if len(path) == 0 {
		// Convert value to proper type before setting
		group.value = p.convertValueToType(value)
		group.isSimple = true
		return
	}

	currentKey := path[0]
	remainingPath := path[1:]

	// Check if currentKey is a number (array index)
	if p.isNumeric(currentKey) {
		// This is an array index
		index, _ := strconv.Atoi(currentKey)

		// Initialize arrayData map if it doesn't exist
		if group.arrayData == nil {
			group.arrayData = make(map[int]*keyGroup)
		}

		if group.arrayData[index] == nil {
			group.arrayData[index] = &keyGroup{
				baseKey:   currentKey,
				children:  make(map[string]*keyGroup),
				arrayData: make(map[int]*keyGroup),
			}
		}
		group.arrayData[index].isArray = true
		p.addNestedToGroup(group.arrayData[index], remainingPath, value)
	} else {
		// This is a regular key
		if group.children == nil {
			group.children = make(map[string]*keyGroup)
		}

		if group.children[currentKey] == nil {
			group.children[currentKey] = &keyGroup{
				baseKey:   currentKey,
				children:  make(map[string]*keyGroup),
				arrayData: make(map[int]*keyGroup),
			}
		}

		child := group.children[currentKey]

		if len(remainingPath) == 0 {
			// This is the final value - convert to proper type
			child.value = p.convertValueToType(value)
			child.isSimple = true
		} else {
			// Continue nesting
			p.addNestedToGroup(child, remainingPath, value)
		}
	}
}

// convertValueToType converts string values to their appropriate types
func (p *Parser) convertValueToType(value string) interface{} {
	// Try to convert to int
	if intVal, err := strconv.Atoi(value); err == nil {
		return intVal
	}

	// Try to convert to int64
	if int64Val, err := strconv.ParseInt(value, 10, 64); err == nil {
		return int64Val
	}

	// Try to convert to float64
	if floatVal, err := strconv.ParseFloat(value, 64); err == nil {
		return floatVal
	}

	// Try to convert to bool
	if boolVal, err := strconv.ParseBool(value); err == nil {
		return boolVal
	}

	// If none of the above, return as string
	return value
}

// buildArrayFromGroup builds an array from a key group
func (p *Parser) buildArrayFromGroup(group *keyGroup) []interface{} {
	if len(group.arrayData) == 0 {
		return []interface{}{}
	}

	// Find max index to determine array size
	maxIndex := 0
	for index := range group.arrayData {
		if index > maxIndex {
			maxIndex = index
		}
	}

	// Create array with proper size
	result := make([]interface{}, maxIndex+1)

	// Process each index
	for index, arrayItem := range group.arrayData {
		if arrayItem.isSimple {
			result[index] = arrayItem.value
		} else if len(arrayItem.children) > 0 || len(arrayItem.arrayData) > 0 {
			// Check if it has children or array data to determine type
			if len(arrayItem.arrayData) > 0 {
				result[index] = p.buildArrayFromGroup(arrayItem)
			} else {
				result[index] = p.buildObjectFromGroup(arrayItem)
			}
		}
	}

	return result
}

// buildObjectFromGroup builds an object from a key group
func (p *Parser) buildObjectFromGroup(group *keyGroup) map[string]interface{} {
	result := make(map[string]interface{})

	// Add simple values
	if group.isSimple {
		result["value"] = group.value
	}

	// Add nested objects
	for key, child := range group.children {
		if child.isSimple {
			result[key] = child.value
		} else if len(child.children) > 0 || len(child.arrayData) > 0 {
			// Check if it has children or array data to determine type
			if len(child.arrayData) > 0 {
				result[key] = p.buildArrayFromGroup(child)
			} else {
				result[key] = p.buildObjectFromGroup(child)
			}
		}
	}

	// Add array data if any - convert int keys to strings
	for key, child := range group.arrayData {
		keyStr := fmt.Sprintf("%d", key) // Convert int to string
		if child.isSimple {
			result[keyStr] = child.value
		} else if len(child.children) > 0 || len(child.arrayData) > 0 {
			// Check if it has children or array data to determine type
			if len(child.arrayData) > 0 {
				result[keyStr] = p.buildArrayFromGroup(child)
			} else {
				result[keyStr] = p.buildObjectFromGroup(child)
			}
		}
	}

	return result
}

// isNumeric checks if a string represents a number
func (p *Parser) isNumeric(s string) bool {
	_, err := strconv.Atoi(s)
	return err == nil
}

// FormToJSONEncoded converts URL-encoded form data with Unicode escapes to JSON
func (p *Parser) FormToJSONEncoded(encodedData string) ([]byte, error) {
	// First, unescape Unicode sequences like \u0026 -> &
	unescapedData := p.unescapeUnicode(encodedData)

	// Then URL decode the data
	decodedData, err := url.QueryUnescape(unescapedData) // Use unescapedData here
	if err != nil {
		return nil, fmt.Errorf("failed to URL decode data: %w", err)
	}

	// Auto-detect format and convert if needed
	formData := p.normalizeFormData(decodedData)

	// Now convert to JSON
	return p.FormToJSON(formData)
}

// convertMultiLineToForm converts multi-line "key = value" format to standard form format
func (p *Parser) convertMultiLineToForm(multiLineData string) string {
	lines := strings.Split(multiLineData, "\n")
	var formParts []string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Split on " = " (space equals space)
		parts := strings.SplitN(line, " = ", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])

			// Convert to standard form format: key=value
			formParts = append(formParts, fmt.Sprintf("%s=%s", key, value))
		}
	}

	// Join with & separator
	return strings.Join(formParts, "&")
}

// FormToJSONEncodedBytes converts URL-encoded form data from bytes to JSON
func (p *Parser) FormToJSONEncodedBytes(data []byte) ([]byte, error) {
	return p.FormToJSONEncoded(string(data))
}

// FormToMapEncoded converts URL-encoded form data with Unicode escapes to a map
func (p *Parser) FormToMapEncoded(encodedData string) (map[string]interface{}, error) {
	// First, unescape Unicode sequences like \u0026 -> &
	unescapedData := p.unescapeUnicode(encodedData)

	// Then URL decode the data
	decodedData, err := url.QueryUnescape(unescapedData)
	if err != nil {
		return nil, fmt.Errorf("failed to URL decode data: %w", err)
	}

	// Now convert to map
	return p.FormToMap(decodedData)
}

// FormToMapEncodedBytes converts URL-encoded form data from bytes to a map
func (p *Parser) FormToMapEncodedBytes(data []byte) (map[string]interface{}, error) {
	return p.FormToMapEncoded(string(data))
}

// unescapeUnicode converts Unicode escape sequences to their actual characters
func (p *Parser) unescapeUnicode(data string) string {
	// Handle common Unicode escapes
	replacements := map[string]string{
		"\\u0026": "&",  // &
		"\\u0027": "'",  // '
		"\\u0022": "\"", // "
		"\\u003C": "<",  // <
		"\\u003E": ">",  // >
		"\\u002B": "+",  // +
		"\\u0020": " ",  // space
	}

	result := data
	for escaped, unescaped := range replacements {
		result = strings.ReplaceAll(result, escaped, unescaped)
	}

	return result
}

// normalizeFormData detects and normalizes different form data formats
func (p *Parser) normalizeFormData(data string) string {
	// Check if it's multi-line format (contains newlines and " = ")
	if strings.Contains(data, "\n") && strings.Contains(data, " = ") {
		return p.convertMultiLineToForm(data)
	}

	// Check if it's standard form format (contains &)
	if strings.Contains(data, "&") {
		return data
	}

	// If it's just one key-value pair
	return data
}
