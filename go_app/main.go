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
	// When running in GUI mode, the application will exit when the GUI window is closed
	// The success message is now handled within the ProcessBooklet function
}
