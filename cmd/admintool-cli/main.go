package main

import (
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
	Subcommands map[string]command
}

var cmds = map[string]command{
	"test": {
		Func: tests.Tester,
		Help: "Run tests [Set NO_INTERACTION environment variable to disable all input interaction]",
	},
	"validate-table": {
		Func: validatetable.ValidateTable,
		Help: "Validate a table",
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
}

func cmdList(cmds map[string]command) {
	fmt.Printf("admintool-cli (commit: %s)\n\n", GitCommit)
	fmt.Println("Commands:")
	for k, cmd := range cmds {
		fmt.Println(k+":", cmd.Help)
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

	if slices.Contains(args, "-h") || slices.Contains(args, "--help") {
		fmt.Printf("usage: %s <command> [args]\n\n", progname)
		cmdList(cmds)
		os.Exit(0)
	}

	cmd, ok := cmds[args[0]]
	if !ok {
		fmt.Printf("unknown command: %s\n\n", args[0])
		cmdList(cmds)
		os.Exit(1)
	}

	if cmd.Subcommands != nil {
		if len(args) < 2 {
			fmt.Printf("usage: %s %s <command> [args]\n%s\n\n", progname, args[0], cmd.Help)
			cmdList(cmd.Subcommands)
			os.Exit(1)
		}

		subcmd, ok := cmd.Subcommands[args[1]]

		if !ok {
			fmt.Printf("unknown command: %s\n\n", args[0]+" "+args[1])
			cmdList(cmd.Subcommands)
			os.Exit(1)
		}

		cmd = subcmd
		args = args[1:]
	}

	cmd.Func(progname, args[1:])
}
