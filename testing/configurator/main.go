package main

import (
	"fmt"

	"github.com/unprofession-al/bpmon"
)

func main() {
	c := bpmon.HealthConfig{}
	err, errs := c.Validate()
	if err != nil {
		fmt.Println(err)
		for _, msg := range errs {
			fmt.Println("- ", msg)
		}
	}
}
