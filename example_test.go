package quuid_test

import (
	"fmt"

	"github.com/MeViksry/quuid"
)

func Example() {
	id := quuid.DeriveID("customers", []byte("customer-42"))
	parsed := quuid.MustParseID(id.String())
	fmt.Println(id.Equal(parsed))
	// Output: true
}
