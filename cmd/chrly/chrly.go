package main

import (
	"fmt"
	"os"

	. "ely.by/chrly/internal/cmd"
)

func main() {
	err := RootCmd.Execute()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
