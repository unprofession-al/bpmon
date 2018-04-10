package main

import (
	"fmt"
	"log"

	"github.com/unprofession-al/bpmon/configuration"
	yaml "gopkg.in/yaml.v2"
)

func main() {
	fmt.Println("--------------")

	err := configuration.Load("/home/daniel/stxt/git/bpmon_config/test.yaml")

	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("--------------")

	c := configuration.GetAll()

	d, err := yaml.Marshal(&c)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	fmt.Printf("--- t dump:\n%s\n\n", string(d))

}
