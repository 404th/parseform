
# ParseForm - Go Library for Form-URLEncoded Data Parsing

## �� Quick Install

```bash
go get github.com/404th/parseform@latest
```

## �� Quick Start

```go
package main

import (
    "fmt"
    "github.com/404th/parseform"
)

func main() {
    parser := parseform.NewParser()
  
    // Simple form to JSON
    formData := "name=John&age=25&email=john@example.com"
    jsonData, err := parser.FormToJSON(formData)
    if err != nil {
        panic(err)
    }
  
    fmt.Println(string(jsonData))
}
```

## 🔐 Multi-line Form Support

```go
// Handle multi-line "key = value" format
multiLineData := `account[id] = 123
account[name] = Example
leads[0][id] = 1`

jsonData, err := parser.FormToJSONEncoded(multiLineData)
```

```

## 4. **Commit and Push README Update**

```bash
git add README.md
git commit -m "Update README with clear installation and usage examples"
git push origin main
```

## 5. **Test Global Installation**

```bash
# In a different directory/project
mkdir test-parseform
cd test-parseform

# Initialize new Go module
go mod init test

# Get your library
go get github.com/404th/parseform@latest

# Check if it was added to go.mod
cat go.mod
```

## 6. **Verify on GitHub**

Make sure these files are visible in your repository:

- ✅ `go.mod`
- ✅ `parseform/parser.go`
- ✅ `README.md`
- ✅ `LICENSE`

## **After These Steps:**

Anyone in the world can use your library with:

```bash
go get github.com/404th/parseform@latest
```

And import it in their code:

```go
import "github.com/404th/parseform"
```

Your library will be globally available! 🚀🌍
