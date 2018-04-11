package main

import (
	"fmt"
	"log"

	"github.com/unprofession-al/bpmon/configuration"
	yaml "gopkg.in/yaml.v2"
)

func main() {

	err := configuration.Load("/home/daniel/stxt/git/bpmon_config/test.yaml")
	if err != nil {
		fmt.Println(err)
	}

	c := configuration.GetSection("default")
	d, err := yaml.Marshal(&c)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	fmt.Printf("%s\n", string(d))

	c.Validate()
}
