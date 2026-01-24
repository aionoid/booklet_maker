package main

import (
	"log"
	"os"
)

func main() {
	cli := &CLI{}
	err := cli.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}

	println("Booklet created successfully!")
}