package main

import (
	"fmt"
)

// Example structs that demonstrate the library usage
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
	ID   int    `form:"name"`
	Name string `form:"name"`
}

type Lead struct {
	ID                int           `form:"id"`
	Name              string        `form:"name"`
	StatusID          int           `form:"status_id"`
	OldStatusID       int           `form:"old_status_id"`
	Price             float64       `form:"price"`
	ResponsibleUserID int           `form:"responsible_user_id"`
	LastModified      int64         `form:"last_modified"`
	ModifiedUserID    int           `form:"modified_user_id"`
	CreatedUserID     int           `form:"created_user_id"`
	DateCreate        int64         `form:"date_create"`
	PipelineID        int           `form:"pipeline_id"`
	Tags              []Tag         `form:"tags"`
	AccountID         int           `form:"account_id"`
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

func main() {
	fmt.Println("ParseForm Library Example")
	fmt.Println("=========================")
	fmt.Println()
	fmt.Println("This is an example of how to use the ParseForm library.")
	fmt.Println("To use it in your project, import it like this:")
	fmt.Println()
	fmt.Println("import \"github.com/404th/parseform\"")
	fmt.Println()
	fmt.Println("Then create a parser and use it:")
	fmt.Println()
	fmt.Println("parser := parseform.NewParser()")
	fmt.Println("var result FormData")
	fmt.Println("err := parser.ParseForm(formData, &result)")
	fmt.Println()
	fmt.Println("See README.md for complete usage examples.")
	fmt.Println()
	fmt.Println("Example structs are defined above for reference.")
}
