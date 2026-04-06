package main

import (
	"fmt"
	"os"
)

var Version = "dev"

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "--version", "-v", "version":
			fmt.Println(Version)
			return
		}
	}

	Run()
}
