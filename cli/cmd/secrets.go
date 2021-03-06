/*
   Copyright 2020 Docker Compose CLI authors

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package cmd

import (
	"fmt"
	"io"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/docker/compose-cli/api/client"
	"github.com/docker/compose-cli/api/secrets"
)

type createSecretOptions struct {
	Label       string
	Username    string
	Password    string
	Description string
}

// SecretCommand manage secrets
func SecretCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "secret",
		Short: "Manages secrets",
	}

	cmd.AddCommand(
		createSecret(),
		inspectSecret(),
		listSecrets(),
		deleteSecret(),
	)
	return cmd
}

func createSecret() *cobra.Command {
	opts := createSecretOptions{}
	cmd := &cobra.Command{
		Use:   "create NAME",
		Short: "Creates a secret.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := client.New(cmd.Context())
			if err != nil {
				return err
			}
			name := args[0]
			secret := secrets.NewSecret(name, opts.Username, opts.Password, opts.Description)
			id, err := c.SecretsService().CreateSecret(cmd.Context(), secret)
			if err != nil {
				return err
			}
			fmt.Println(id)
			return nil
		},
	}

	cmd.Flags().StringVarP(&opts.Username, "username", "u", "", "username")
	cmd.Flags().StringVarP(&opts.Password, "password", "p", "", "password")
	cmd.Flags().StringVarP(&opts.Description, "description", "d", "", "Secret description")
	return cmd
}

func inspectSecret() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "inspect ID",
		Short: "Displays secret details",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := client.New(cmd.Context())
			if err != nil {
				return err
			}
			secret, err := c.SecretsService().InspectSecret(cmd.Context(), args[0])
			if err != nil {
				return err
			}
			out, err := secret.ToJSON()
			if err != nil {
				return err
			}
			fmt.Println(out)
			return nil
		},
	}
	return cmd
}

func listSecrets() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List secrets stored for the existing account.",
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := client.New(cmd.Context())
			if err != nil {
				return err
			}
			list, err := c.SecretsService().ListSecrets(cmd.Context())
			if err != nil {
				return err
			}
			printList(os.Stdout, list)
			return nil
		},
	}
	return cmd
}

type deleteSecretOptions struct {
	recover bool
}

func deleteSecret() *cobra.Command {
	opts := deleteSecretOptions{}
	cmd := &cobra.Command{
		Use:     "delete NAME",
		Aliases: []string{"rm", "remove"},
		Short:   "Removes a secret.",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := client.New(cmd.Context())
			if err != nil {
				return err
			}
			return c.SecretsService().DeleteSecret(cmd.Context(), args[0], opts.recover)
		},
	}
	cmd.Flags().BoolVar(&opts.recover, "recover", false, "Enable recovery.")
	return cmd
}

func printList(out io.Writer, secrets []secrets.Secret) {
	printSection(out, func(w io.Writer) {
		for _, secret := range secrets {
			fmt.Fprintf(w, "%s\t%s\t%s\n", secret.ID, secret.Name, secret.Description) // nolint:errcheck
		}
	}, "ID", "NAME", "DESCRIPTION")
}

func printSection(out io.Writer, printer func(io.Writer), headers ...string) {
	w := tabwriter.NewWriter(out, 20, 1, 3, ' ', 0)
	fmt.Fprintln(w, strings.Join(headers, "\t")) // nolint:errcheck
	printer(w)
	w.Flush() // nolint:errcheck
}
