# go-influitive

## example

```go
package main

import (
    "fmt"
    "os"

    influitive "github.com/hokiegeek/go-influitive"
)

func main() {
    influitiveToken := os.Getenv("INFLUITIVE_API_TOKEN")
    influitiveOrgID := os.Getenv("INFLUITIVE_ORG_ID")

    if len(influitiveToken) == 0 || len(influitiveOrgID) == 0 {
        panic("Environment variables not set")
    }

    client, err := influitive.NewClient(influitiveToken, influitiveOrgID)
    if err != nil {
        panic(err)
    }

    newMember, err := influitive.CreateMemberByEmail(client, "sdood@exam.ple", "Sohm Dood", "example")
    if err != nil {
        panic(err)
    }

    if err := influitive.InviteMember(client, newMember.ID, false); err != nil {
        panic(err)
    }

    members, err := influitive.GetAllMembers(client)
    if err != nil {
        panic(err)
    }

    fmt.Printf("retrieved %d members\n", len(members))
    for _, m := range members {
        fmt.Println(m.Email)
    }
}
```
