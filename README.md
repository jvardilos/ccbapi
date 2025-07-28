# CCB API Handler

This Go package wraps the **Church Community Builder (CCB) v2 API** using **OAuth 2.0** authentication. It simplifies everything from initial login to making authenticated API calls — all with automatic background token refreshing.

## Features

- Handles the full **OAuth 2.0 Authorization Code Flow**
- Automatically refreshes access tokens when needed
- Returns raw `[]byte` responses — ready to print or `json.Unmarshal`
- No need to manage headers, encoding, or query strings manually

## Installation

```bash
go get github.com/jvardilos/ccbapi@latest
```

## Sample Implementation

```go
package main

import (
	"fmt"
	"github.com/jvardilos/ccbapi"
)

func CCB() error {
	c := ccbapi.Credentials{
		Subdomain: "yourchurch",
		Client:    "your_client_id",
		Secret:    "your_client_secret",
	}

	t, err := ccbapi.Auth(&c)
	if err != nil {
		return err
	}

	body, err := ccbapi.Call("GET", "individuals", t, &c)
	if err != nil {
		return err
	}

	fmt.Println(string(body)) // or json.Unmarshal if you'd like
	return nil
}
```

## API Resources

- [API Docs](https://api.ccbchurch.com/documentation)
- [Swagger Spec (YAML)](https://api.ccbchurch.com/documentation.yaml)
- [Request Client ID & Secret](https://vendor.ccbchurch.com/goto/forms/15/responses/new)
