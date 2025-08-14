# ParseForm - Go Library for Form-URLEncoded Data Parsing

A powerful and flexible Go library for parsing form-urlencoded data into JSON, Go maps, or structs. Supports nested structures, arrays, automatic type conversion, and handles various form data formats including URL-encoded data with Unicode escapes.

## ✨ Features

- 🚀 **Dynamic parsing** - No predefined structs required
- 🔗 **Nested structures** - Handle complex nested form data
- 📚 **Array support** - Automatically detect and build arrays
- ️ **Type conversion** - Automatic conversion to int, float, bool, string
- 🌍 **Multi-format support** - Standard form, multi-line, URL-encoded, Unicode escapes
- ⚡ **High performance** - Efficient parsing with minimal overhead
- 🔧 **Flexible output** - JSON, Go maps, or structs

## 📦 Installation

```bash
go get github.com/404th/parseform@latest
```

## 🚀 Quick Start

### Basic Usage

```go
package main

import (
    "fmt"
    "github.com/404th/parseform"
)

func main() {
    parser := parseform.NewParser()
  
    // Simple form data to JSON
    formData := "name=John&age=25&email=john@example.com"
    jsonData, err := parser.FormToJSON(formData)
    if err != nil {
        panic(err)
    }
  
    fmt.Println(string(jsonData))
}
```

## API Reference

### Core Methods

#### Form to JSON

```go
// Convert form data to JSON
jsonData, err := parser.FormToJSON("name=John&age=25")

// Convert form data from bytes to JSON
jsonData, err := parser.FormToJSONBytes([]byte("name=John&age=25"))

// Convert URL-encoded data with Unicode escapes to JSON
jsonData, err := parser.FormToJSONEncoded("account%5Bid%5D=123\u0026name=John")

// Convert URL-encoded bytes to JSON
jsonData, err := parser.FormToJSONEncodedBytes([]byte("account%5Bid%5D=123"))
```

#### Form to Go Maps

```go
// Convert form data to Go map
resultMap, err := parser.FormToMap("name=John&age=25")

// Convert form data from bytes to Go map
resultMap, err := parser.FormToMapBytes([]byte("name=John&age=25"))

// Convert URL-encoded data to Go map
resultMap, err := parser.FormToMapEncoded("account%5Bid%5D=123\u0026name=John")

// Convert URL-encoded bytes to Go map
resultMap, err := parser.FormToMapEncodedBytes([]byte("account%5Bid%5D=123"))
```

#### Struct Parsing (Traditional)

```go
type User struct {
    Name string `form:"name"`
    Age  int    `form:"age"`
}

var user User
err := parser.ParseForm("name=John&age=25", &user)
```

## 🔐 Supported Form Data Formats

### 1. Standard Form Data

```
name=John&age=25&email=john@example.com
```

### 2. Nested Objects

```
user[name]=John&user[profile][age]=25&user[profile][city]=NY
```

### 3. Arrays

```
items[0][id]=1&items[0][name]=Item1&items[1][id]=2&items[1][name]=Item2
```

### 4. Mixed Complex Structures

```
account[subdomain]=example&account[users][0][id]=1&account[users][0][permissions][read]=true
```

### 5. Multi-line Format

```
account[id] = 123
account[name] = Example
leads[0][id] = 1
leads[0][status] = active
```

### 6. URL-encoded Data

```
account%5Bid%5D=123&name%3DJohn
```

### 7. Unicode Escapes

```
account\u0026id=123&name\u003DJohn
```

## 🎯 Real-World Examples

### CRM Lead Data

```go
// Complex CRM data with nested structures
crmData := `account[subdomain]=example&account[id]=123&leads[0][id]=1&leads[0][name]=Lead1&leads[0][tags][0]=urgent&leads[0][custom_fields][0][id]=100&leads[0][custom_fields][0][value]=Important`

parser := parseform.NewParser()
jsonData, err := parser.FormToJSON(crmData)
if err != nil {
    panic(err)
}

fmt.Println(string(jsonData))
```

### Multi-line Configuration

```go
// Configuration data in multi-line format
configData := `database[host] = localhost
database[port] = 5432
database[ssl] = true
redis[host] = 127.0.0.1
redis[port] = 6379`

parser := parseform.NewParser()
resultMap, err := parser.FormToMapEncoded(configData)
if err != nil {
    panic(err)
}

fmt.Printf("%+v\n", resultMap)
```

### Webhook Data Processing

```go
// Handle webhook data with URL encoding
webhookData := `user%5Bid%5D=123&user%5Bname%5D=John&events%5B0%5D%5Btype%5D=login&events%5B0%5D%5Btimestamp%5D=1640995200`

parser := parseform.NewParser()
jsonData, err := parser.FormToJSONEncoded(webhookData)
if err != nil {
    panic(err)
}

fmt.Println(string(jsonData))
```

## 🔧 Advanced Usage

### Custom Type Handling

The parser automatically converts values to appropriate types:

```go
// Input: "age=25&price=99.99&active=true&name=John"
// Output types:
// - age: int (25)
// - price: float64 (99.99)
// - active: bool (true)
// - name: string ("John")
```

### Error Handling

```go
jsonData, err := parser.FormToJSON(formData)
if err != nil {
    switch {
    case strings.Contains(err.Error(), "failed to parse form data"):
        log.Printf("Form data parsing error: %v", err)
    case strings.Contains(err.Error(), "failed to marshal to JSON"):
        log.Printf("JSON marshaling error: %v", err)
    default:
        log.Printf("Unknown error: %v", err)
    }
    return
}
```

## Requirements

- Go 1.21 or higher
- No external dependencies

## Performance

- **Efficient parsing** - Uses Go's built-in `url.ParseQuery`
- **Smart type detection** - Minimal overhead for type conversion
- **Memory optimized** - Efficient data structures for large forms

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- Built with Go's powerful reflection and URL parsing capabilities
- Inspired by the need for flexible form parsing in web applications
- Tested with real-world CRM, webhook, and configuration data

## 📞 Support

If you have any questions or need help, please open an issue on GitHub.

---

**Made with ❤️ in Go by 404th**
