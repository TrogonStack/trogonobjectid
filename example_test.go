package trogonobjectid_test

import (
	"encoding/json"
	"fmt"

	"github.com/TrogonStack/trogonobjectid"
)

func Example() {
	id := trogonobjectid.MustNew[invoiceType]("0199e3a1-7c2b-7d4e-9f10-2a3b4c5d6e7f")

	data, _ := json.Marshal(id)
	fmt.Println(id)
	fmt.Println(id.Storage())
	fmt.Println(string(data))
	// Output:
	// inv_0199e3a1-7c2b-7d4e-9f10-2a3b4c5d6e7f
	// 0199e3a1-7c2b-7d4e-9f10-2a3b4c5d6e7f
	// "0199e3a1-7c2b-7d4e-9f10-2a3b4c5d6e7f"
}

func ExampleParse() {
	id, err := trogonobjectid.Parse[userType]("user_abc-123")
	if err != nil {
		panic(err)
	}
	fmt.Println(id.Bare())

	_, err = trogonobjectid.Parse[userType]("order_abc-123")
	fmt.Println(err)
	// Output:
	// abc-123
	// trogonobjectid: invalid user value "order_abc-123": invalid format
}
