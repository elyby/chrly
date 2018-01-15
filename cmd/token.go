package cmd

import (
	"fmt"
	"log"

	"elyby/minecraft-skinsystem/auth"

	"github.com/segmentio/go-prompt"
	"github.com/spf13/cobra"
)

var tokenCmd = &cobra.Command{
	Use:   "token",
	Short: "API tokens operations",
}

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create the new token, that allows interacting with Ely.by Skinsystem API",
	Run: func(cmd *cobra.Command, args []string) {
		jwtAuth := &auth.JwtAuth{}
		for {
			token, err := jwtAuth.NewToken(auth.SkinScope)
			if err != nil {
				if _, ok := err.(*auth.SigningKeyNotAvailable); !ok {
					log.Fatalf("Unable to create new token. The error is %v\n", err)
				}

				log.Println("Signing key not available. Creating...")
				err := jwtAuth.GenerateSigningKey()
				if err != nil {
					log.Fatalf("Unable to generate new signing key. The error is %v\n", err)
				}

				continue
			}

			fmt.Printf("%s\n", token)
		}
	},
}

var resetCmd = &cobra.Command{
	Use:   "reset",
	Short: "Regenerate the secret key, that invalidate all tokens",
	Run: func(cmd *cobra.Command, args []string) {
		if !prompt.Confirm("Do you really want to invalidate all exists tokens?") {
			fmt.Println("Aboart.")
			return
		}

		jwtAuth := &auth.JwtAuth{}
		if err := jwtAuth.GenerateSigningKey(); err != nil {
			log.Fatalf("Unable to generate new signing key. The error is %v\n", err)
		}

		fmt.Println("Token successfully regenerated.")
	},
}

func init() {
	tokenCmd.AddCommand(createCmd, resetCmd)
	RootCmd.AddCommand(tokenCmd)
}
