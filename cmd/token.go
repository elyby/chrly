package cmd

import (
	"fmt"
	"log"

	"github.com/elyby/chrly/http"

	"github.com/spf13/cobra"
)

var tokenCmd = &cobra.Command{
	Use:   "token",
	Short: "Creates a new token, which allows to interact with Chrly API",
	Run: func(cmd *cobra.Command, args []string) {
		container := shouldGetContainer()
		var auth *http.JwtAuth
		err := container.Resolve(&auth)
		if err != nil {
			log.Fatal(err)
		}

		token, err := auth.NewToken(http.SkinScope)
		if err != nil {
			log.Fatalf("Unable to create new token. The error is %v\n", err)
		}

		fmt.Printf("%s\n", token)
	},
}

func init() {
	RootCmd.AddCommand(tokenCmd)
}
