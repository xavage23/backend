package main

import (
	"admintool-cli/bulkstock"
	"admintool-cli/tests"
	"admintool-cli/users/adduser"
	"admintool-cli/users/createmultipleusers"
	"admintool-cli/validatetable"
	"fmt"

	cmd "github.com/infinitybotlist/eureka/cmd"
)

func main() {
	state := cmd.CommandLineState{
		Commands: map[string]cmd.Command{
			"test": {
				Func: tests.Tester,
				Help: "Run tests [Set NO_INTERACTION environment variable to disable all input interaction]",
			},
			"validate-table": {
				Func:    validatetable.ValidateTable,
				Help:    "Validate a table",
				Usage:   "validate-table <database> <target/ref_column> <backer/column>",
				Example: "validate-table infinity reviews/author users/user_id",
				ArgValidate: func(args []string) error {
					if len(args) != 3 {
						return fmt.Errorf("expected 3 arguments, got %d", len(args))
					}

					return nil
				},
			},
			"users": {
				Help: "Manage users",
				Subcommands: map[string]cmd.Command{
					"create": {
						Func: adduser.CreateUser,
						Help: "Create a user",
					},
					"createmultiple": {
						Func: createmultipleusers.CreateMultipleUsers,
						Help: "Create multiple users based on a list of newline-seperated usernames ending with an EOF",
					},
				},
			},
			"bulkstock": {
				Help: "Bulk import stock data",
				Subcommands: map[string]cmd.Command{
					"import": {
						Func:    bulkstock.BulkImportStock,
						Help:    "Import stock data from a YAML file",
						Usage:   "bulkstock import <database> <file>",
						Example: "bulkstock import xavage output.yaml",
						ArgValidate: func(args []string) error {
							if len(args) != 2 {
								return fmt.Errorf("expected 2 argument, got %d", len(args))
							}

							return nil
						},
					},
					"showpc": {
						Func:    bulkstock.ShowPriceChanges,
						Help:    "Show price changes from a YAML file",
						Usage:   "bulkstock showpc <file>",
						Example: "bulkstock showpc output.yaml",
						ArgValidate: func(args []string) error {
							if len(args) != 1 {
								return fmt.Errorf("expected 1 argument, got %d", len(args))
							}

							return nil
						},
					},
				},
			},
		},
		GetHeader: func() string {
			return fmt.Sprintf("admintool-cli %s", cmd.GetGitCommit())
		},
	}

	state.Run()
}
