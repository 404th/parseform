# ParseForm - Go Library for Form-URLEncoded Data Parsing

A powerful and flexible Go library for parsing form-urlencoded data into struct types. Supports nested structures, arrays, maps, and automatic type conversion.

## ✨ Features

- 🚀 **Automatic parsing** - Parse form data directly into Go structs
- 🔗 **Nested structures** - Handle complex nested form data with ease
- 📚 **Array support** - Automatically parse indexed arrays and slices
- 🗺️ **Map support** - Parse form data into Go maps
- 🏷️ **Tag-based mapping** - Use `form` tags to map form fields to struct fields
- 🔧 **Type conversion** - Automatic conversion between string and Go types
- 🌍 **UTF-8 support** - Handle international characters correctly
- ⚡ **High performance** - Reflection-based parsing with minimal overhead

## 📦 Installation

```bash
go get github.com/404th/parseform
```

Or clone from GitHub:

```bash
git clone https://github.com/404th/parseform.git
cd parseform
go mod tidy
```

## 🚀 Quick Start

### Basic Usage

```go
package main

import (
    "fmt"
    "log"
    "github.com/404th/parseform"
)

type User struct {
    Name  string `form:"name"`
    Email string `form:"email"`
    Age   int    `form:"age"`
}

func main() {
    formData := "name=John&email=john@example.com&age=25"
  
    parser := parseform.NewParser()
    var user User
  
    err := parser.ParseForm(formData, &user)
    if err != nil {
        log.Fatalf("Error parsing form: %v", err)
    }
  
    fmt.Printf("User: %+v\n", user)
}
```

### Nested Structures

```go
type Account struct {
    ID       int    `form:"id"`
    Username string `form:"username"`
    Profile  struct {
        FirstName string `form:"first_name"`
        LastName  string `form:"last_name"`
    } `form:"profile"`
}

// Form data: account[id]=123&account[username]=john&account[profile][first_name]=John&account[profile][last_name]=Doe
```

### Arrays and Slices

```go
type Lead struct {
    ID     int      `form:"id"`
    Name   string   `form:"name"`
    Tags   []string `form:"tags"`
}

// Form data: leads[0][id]=1&leads[0][name]=Lead1&leads[0][tags][0]=urgent&leads[0][tags][1]=new
```

## 📚 Advanced Examples

### Complex CRM Data Structure

```go
type Account struct {
    Subdomain string `form:"subdomain"`
    ID        int    `form:"id"`
    Links     struct {
        Self string `form:"self"`
    } `form:"_links"`
}

type CustomField struct {
    ID     int    `form:"id"`
    Name   string `form:"name"`
    Values []struct {
        Value string `form:"value"`
        Enum  int    `form:"enum"`
    } `form:"values"`
    Code string `form:"code"`
}

type Tag struct {
    ID   int    `form:"id"`
    Name string `form:"name"`
}

type Lead struct {
    ID                int           `form:"id"`
    Name              string        `form:"name"`
    StatusID          int           `form:"status_id"`
    Price             float64       `form:"price"`
    Tags              []Tag         `form:"tags"`
    CustomFields      []CustomField `form:"custom_fields"`
    CreatedAt         int64         `form:"created_at"`
    UpdatedAt         int64         `form:"updated_at"`
}

type FormData struct {
    Account Account `form:"account"`
    Leads   struct {
        Status []Lead `form:"status"`
    } `form:"leads"`
}
```

### HTTP Handler Example

```go
func handleFormSubmission(w http.ResponseWriter, r *http.Request) {
    if err := r.ParseForm(); err != nil {
        http.Error(w, "Failed to parse form", http.StatusBadRequest)
        return
    }
  
    // Convert form data to string for parsing
    formData := r.Form.Encode()
  
    var data FormData
    parser := parseform.NewParser()
  
    if err := parser.ParseForm(formData, &data); err != nil {
        http.Error(w, "Failed to parse form data", http.StatusBadRequest)
        return
    }
  
    // Process the parsed data
    fmt.Printf("Account: %s (ID: %d)\n", data.Account.Subdomain, data.Account.ID)
  
    w.WriteHeader(http.StatusOK)
    w.Write([]byte("Form processed successfully"))
}
```

## 🔧 API Reference

### Parser

- `NewParser() *Parser` - Create a new parser instance
- `ParseForm(formData string, target interface{}) error` - Parse form data into a struct
- `ParseFormBytes(data []byte, target interface{}) error` - Parse form data from bytes

### Supported Types

- `string`
- `int`, `int8`, `int16`, `int32`, `int64`
- `uint`, `uint8`, `uint16`, `uint32`, `uint64`
- `float32`, `float64`
- `bool`
- `struct` (nested structures)
- `slice` (arrays)
- `map` (maps)

### Utility Functions

- `ParseTimestamp(timestamp string) (int64, error)` - Parse Unix timestamp
- `ParseInt(value string) (int, error)` - Parse string to int
- `ParseFloat(value string) (float64, error)` - Parse string to float64

## 🏷️ Struct Tags

Use the `form` tag to map form field names to struct fields:

```go
type User struct {
    ID        int    `form:"user_id"`           // Maps user_id to ID
    FirstName string `form:"first_name"`        // Maps first_name to FirstName
    LastName  string `form:"last_name"`         // Maps last_name to LastName
}
```

## 🔍 Form Data Format

The library supports various form data formats:

### Simple Fields

```
name=John&email=john@example.com&age=25
```

### Nested Fields

```
user[name]=John&user[profile][age]=25&user[profile][city]=New York
```

### Arrays

```
items[0][id]=1&items[0][name]=Item1&items[1][id]=2&items[1][name]=Item2
```

### Mixed Complexity

```
account[subdomain]=example&account[id]=123&leads[0][id]=1&leads[0][tags][0]=urgent
```

## 🚨 Error Handling

Always check for errors when parsing:

```go
err := parser.ParseForm(formData, &result)
if err != nil {
    switch {
    case strings.Contains(err.Error(), "failed to parse form data"):
        log.Printf("Form data parsing error: %v", err)
    case strings.Contains(err.Error(), "failed to parse field"):
        log.Printf("Field parsing error: %v", err)
    default:
        log.Printf("Unknown error: %v", err)
    }
    return
}
```

## 📋 Requirements

- Go 1.21 or higher
- No external dependencies

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- Built with Go's powerful reflection capabilities
- Inspired by the need for flexible form parsing in web applications
- Tested with real-world CRM and web form data

## 📞 Support

If you have any questions or need help, please open an issue on GitHub.

---

**Made with ❤️ in Go by 404th**
