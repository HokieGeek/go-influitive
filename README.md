# go-influitive

[![GoDoc](https://pkg.go.dev/badge/github.com/HokieGeek/go-influitive?status.svg)](https://pkg.go.dev/github.com/HokieGeek/go-influitive?tab=doc)

A quick and dirty implementation of some calls from the Influitive API (<https://influitive.readme.io/reference>) used for testing and discovery when implementing integrations.

## example usage

```go
package main

import (
    "fmt"
    "os"

    influitive "github.com/hokiegeek/go-influitive"
)

func main() {
    // Need to retrieve the API token and Org ID from the Influitive 'Integrations' settings
    influitiveToken := os.Getenv("INFLUITIVE_API_TOKEN")
    influitiveOrgID := os.Getenv("INFLUITIVE_ORG_ID")

    if len(influitiveToken) == 0 || len(influitiveOrgID) == 0 {
        panic("Environment variables not set")
    }

    // Create a client to Influitive
    client, err := influitive.NewClient(influitiveToken, influitiveOrgID)
    if err != nil {
        panic(err)
    }

    // Create a new uninvited member
    newMember, err := influitive.CreateMemberByEmail(client, "sdood@exam.ple", "Sohm Dood", "example")
    if err != nil {
        panic(err)
    }

    // Invite the newly created member
    if err := influitive.InviteMember(client, newMember.ID, false); err != nil {
        panic(err)
    }

    // Retrieve all of the members in your Organization
    members, err := influitive.GetAllMembers(client)
    if err != nil {
        panic(err)
    }

    // .... and list their emails
    fmt.Printf("retrieved %d members\n", len(members))
    for _, m := range members {
        fmt.Println(m.Email)
    }
}
```
