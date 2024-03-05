package cmd

import (
	"fmt"

	"ely.by/chrly/internal/security"

	"github.com/spf13/cobra"
)

var tokenCmd = &cobra.Command{
	Use:       "token scope1 ...",
	Example:   "token profiles sign",
	Short:     "Creates a new token, which allows to interact with Chrly API",
	ValidArgs: []string{string(security.ProfilesScope), string(security.SignScope)},
	RunE: func(cmd *cobra.Command, args []string) error {
		container := shouldGetContainer()
		var auth *security.Jwt
		err := container.Resolve(&auth)
		if err != nil {
			return err
		}

		scopes := make([]security.Scope, len(args))
		for i := range args {
			scopes[i] = security.Scope(args[i])
		}

		token, err := auth.NewToken(scopes...)
		if err != nil {
			return fmt.Errorf("Unable to create a new token. The error is %v\n", err)
		}

		fmt.Println(token)

		return nil
	},
}

func init() {
	RootCmd.AddCommand(tokenCmd)
}
