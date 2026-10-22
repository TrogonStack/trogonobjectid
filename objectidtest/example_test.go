package objectidtest_test

import (
	"fmt"

	"github.com/TrogonStack/trogonobjectid/objectidtest"
)

func ExampleFromSeed() {
	buyer := objectidtest.FromSeed[fanType]("buyer")
	fmt.Println(buyer == objectidtest.FromSeed[fanType]("buyer"))
	fmt.Println(buyer.Prefix())
	// Output:
	// true
	// fan_
}
