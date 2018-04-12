package main

import (
	"fmt"
	"log"

	"github.com/unprofession-al/bpmon/configuration"
	yaml "gopkg.in/yaml.v2"
)

func main() {
	c, err := configuration.Load("/home/daniel/stxt/git/bpmon_config/test.yaml")
	if err != nil {
		fmt.Println(err)
	}

	s := c.Section("default")
	d, err := yaml.Marshal(&s)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	fmt.Printf("%s\n", string(d))

	_, errs := s.Validate()
	for _, i := range errs {
		fmt.Println(i)
	}
}
