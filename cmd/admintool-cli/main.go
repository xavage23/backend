package main

import (
	"admintool-cli/bulkimport"
	"admintool-cli/tests"
	"admintool-cli/users/adduser"
	"admintool-cli/validatetable"
	"fmt"
	"os"
	"runtime/debug"

	"golang.org/x/exp/slices"
)

var GitCommit string

func init() {
	// Use runtime/debug vcs.revision to get the git commit hash
	if info, ok := debug.ReadBuildInfo(); ok {
		for _, setting := range info.Settings {
			if setting.Key == "vcs.revision" {
				GitCommit = setting.Value
			}
		}
	}

	if GitCommit == "" {
		GitCommit = "unknown"
	}
}

type command struct {
	Func        func(progname string, args []string)
	Help        string
	Usage       string
	Example     string
	Subcommands map[string]command
	ArgValidate func(args []string) error
}

func (c *command) Validate(args []string) error {
	if c.ArgValidate != nil {
		err := c.ArgValidate(args)

		if err != nil {
			return fmt.Errorf("invalid arguments: %w", err)
		}
	}

	return nil
}

func FindCommand(args []string) (*command, []string, error) {
	if len(args) == 0 {
		return nil, args, fmt.Errorf("no command provided")
	}

	c, ok := cmds[args[0]]
	if !ok {
		return nil, args, fmt.Errorf("unknown command: %s", args[0])
	}

	if c.Subcommands != nil {
		if len(args) < 2 {
			if c.Func != nil {
				return &c, args, nil
			}

			return &c, args, fmt.Errorf("no subcommand provided")
		}

		subcmd, ok := c.Subcommands[args[1]]

		if !ok {
			return &c, args, fmt.Errorf("unknown subcommand: %s", args[0]+" "+args[1])
		}

		c = subcmd
		args = args[2:]
	} else {
		args = args[1:]
	}

	return &c, args, nil
}

func (c *command) GetUsage() string {
	initial := c.Help

	if c.Usage != "" {
		initial += "\n\nUsage: " + c.Usage
	}

	if c.Example != "" {
		initial += "\n\nExample: " + c.Example
	}

	if c.Subcommands != nil {
		initial += "\n\nSubcommands:"

		for k, cmd := range c.Subcommands {
			initial += fmt.Sprintf("\n%s: %s", k, cmd.Help)
		}
	}

	return initial
}

var cmds = map[string]command{
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
		Subcommands: map[string]command{
			"create": {
				Func: adduser.CreateUser,
				Help: "Create a user",
			},
		},
	},
	"bulkimport": {
		Func:    bulkimport.BulkImport,
		Help:    "Bulk import data",
		Usage:   "bulkimport <database> <file>",
		Example: "bulkimport xavage import.yaml",
		ArgValidate: func(args []string) error {
			if len(args) != 2 {
				return fmt.Errorf("expected 2 argument, got %d", len(args))
			}

			return nil
		},
	},
}

func cmdListToArray(cmds map[string]command) []string {
	s := []string{"Commands:"}
	for k, cmd := range cmds {
		s = append(s, fmt.Sprint(k+": ", cmd.Help))
	}

	return s
}

func cmdList(cmds map[string]command) {
	for _, cmd := range cmdListToArray(cmds) {
		fmt.Println(cmd)
	}
}

func main() {
	progname := os.Args[0]
	args := os.Args[1:]

	if len(args) == 0 {
		fmt.Printf("usage: %s <command> [args]\n\n", progname)
		cmdList(cmds)
		os.Exit(1)
	}

	cmd, args, err := FindCommand(args)

	if slices.Contains(args, "-h") || slices.Contains(args, "--help") {
		fmt.Printf("admintool-cli (commit: %s)\n", GitCommit)
		fmt.Printf("structure: %s <command> [args]\n\n", progname)

		if cmd != nil {
			fmt.Printf("%s\n\n", cmd.GetUsage())
		} else {
			cmdList(cmds)
		}

		os.Exit(1)
	}

	if err != nil {
		fmt.Printf("error: %s\n\n", err)

		if cmd != nil {
			fmt.Printf("structure: %s [args]\n%s\n\n", progname, cmd.GetUsage())
		} else {
			cmdList(cmds)
		}

		os.Exit(1)
	}

	if err := cmd.Validate(args); err != nil {
		fmt.Printf("error: %s\n\n", err)
		fmt.Printf("structure: %s [args]\n%s\n\n", progname, cmd.GetUsage())
		os.Exit(1)
	}

	cmd.Func(progname, args)
}
