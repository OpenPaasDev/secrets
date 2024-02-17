package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/openpaasDev/secrets/pkg/secrets"
	"github.com/spf13/cobra"
)

func main() {
	// Do something
	rootCmd := &cobra.Command{
		Use:   "secrets",
		Short: "secrets allows you to manage secrets for your environment",
		Long:  `secrets - A tool to manage secrets for your environment`,
		Run: func(cmd *cobra.Command, args []string) {
			err := cmd.Help()
			if err != nil {
				panic(err)
			}
		},
	}

	subCommands := initSecretsCommand()
	for _, cmd := range subCommands {
		rootCmd.AddCommand(cmd)
	}
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func initSecretsCommand() []*cobra.Command {
	var environment string
	var baseDir string
	var name string
	var outputFile string
	subCommands := []*cobra.Command{
		{
			Use:   "init",
			Short: "initialize secret management for environment",
			Long:  `initialize secret management for environment`,
			Run: func(cmd *cobra.Command, args []string) {
				initSecretSystem(baseDir, environment)
			},
		},
		{
			Use:   "add",
			Short: "add a new secret",
			Long:  `add a new secret`,
			Run: func(cmd *cobra.Command, args []string) {
				err := secrets.AddSecret(baseDir, environment, name)
				if err != nil {
					checkSecretInit(err, environment)
					fmt.Println(err)
					os.Exit(1)
				}
			},
		},
		{
			Use:   "env",
			Short: "dump all secrets as `export KEY=value` into the desired output file (-o)",
			Long:  "dump all secrets as `export KEY=value` into the desired output file (-o)",
			Run: func(cmd *cobra.Command, args []string) {
				homeDir, err := os.UserHomeDir()
				if err != nil {
					fmt.Println(err)
					os.Exit(1)
				}
				secrets, err := secrets.GetAllSecrets(homeDir, baseDir, environment)
				if err != nil {
					checkSecretInit(err, environment)
					fmt.Println(err)
					os.Exit(1)
				}

				envrc := ""
				for _, secret := range secrets {
					envrc += fmt.Sprintf("export %s=%s\n", secret.Name, secret.Value)
				}
				err = os.WriteFile(outputFile, []byte(envrc), 0644) // nolint
				if err != nil {
					log.Fatalf("Failed to write to file: %v", err)
				}
			},
		},

		{
			Use:   "refresh",
			Short: "re-encrypts all secrets with currently available public keys",
			Long:  `re-encrypts all secrets with currently available public keys, useful when new keys are added or removed`,
			Run: func(cmd *cobra.Command, args []string) {
				dirname, err := os.UserHomeDir()
				if err != nil {
					fmt.Println(err)
					os.Exit(1)
				}
				err = secrets.Refresh(dirname, baseDir, environment)
				if err != nil {
					checkSecretInit(err, environment)
					fmt.Println(err)
					os.Exit(1)
				}
			},
		},
	}
	for _, cmd := range subCommands {
		cmd.Flags().StringVarP(&baseDir, "baseDir", "b", "", "base directory for secrets")
		err := cmd.MarkFlagRequired("baseDir")
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		cmd.Flags().StringVarP(&environment, "environment", "e", "", "Environment name")
		err = cmd.MarkFlagRequired("environment")
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		if cmd.Name() == "add" {
			cmd.Flags().StringVarP(&name, "name", "n", "", "Name of secret")
			err = cmd.MarkFlagRequired("name")
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		}
		if cmd.Name() == "env" {
			cmd.Flags().StringVarP(&outputFile, "output", "o", "", "Name of output file")
			err = cmd.MarkFlagRequired("output")
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		}
	}

	return subCommands
}

func initSecretSystem(baseDir, environment string) {
	dirname, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}
	initialized, err := secrets.InitSecrets(dirname, baseDir, environment)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	if initialized {
		fmt.Println("")
		fmt.Println("A private key has been created for you in $HOME/.openpaas/private-key.asc")
		fmt.Println("Please make sure to keep your private key and passphrase secure from prying eyes and backed up.")
		fmt.Println("A lost private key & passphrase cannot be recovered. If no user remains with access to at least one functioning key, secrets data will be lost.")
		fmt.Println("")
		fmt.Println("Keys & passphrases should not be shared. Each user requiring access should create their own with `secrets init`, then commit their public key to the repository and request an existing user to re-encrypt all secrets with `secrets refresh`")
	}
}

func checkSecretInit(err error, environment string) {
	if strings.Contains(err.Error(), "private-key.asc: no such file or directory") {
		fmt.Printf("No private key found. Have you previously run `secrets init -e %s` to initialise the secrets for the environment?", environment)
		fmt.Println("")
		fmt.Println("If you have run `secrets init`, but secrets exist from before, they may need to be re-encrypted by a colleague with `secrets refresh`")
		os.Exit(1)
	}
}
