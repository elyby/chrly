package main

import (
	"fmt"
	"os"

	. "github.com/elyby/chrly/internal/cmd"
)

func main() {
	err := RootCmd.Execute()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
