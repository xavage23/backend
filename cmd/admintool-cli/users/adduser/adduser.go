package adduser

import (
	"admintool-cli/common"
	"fmt"
	"xavagebb/state"

	"github.com/alexedwards/argon2id"
	"github.com/infinitybotlist/eureka/crypto"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/exp/slices"
)

func CreateUser(progname string, args []string) {
	var isAdmin bool

	if slices.Contains(args, "-a") || slices.Contains(args, "--admin") {
		isAdmin = true
	}

	var name string

	for _, arg := range args {
		if arg != "-a" || arg == "--admin" {
			name = arg
			break
		}
	}

	if name == "" {
		common.Fatal("No name provided")
	}

	fmt.Println("Got name:", name)
	fmt.Println("Is admin:", isAdmin)

	for {
		check := common.AskInput("Is this correct? [y/n] ")

		if check == "y" || check == "Y" {
			break
		}

		if check == "n" || check == "N" {
			common.Fatal("Aborting...")
		}
	}

	fmt.Println("Creating user...")

	initialPw := crypto.RandString(16)
	token := crypto.RandString(512)

	argon2hash, err := argon2id.CreateHash(initialPw, argon2id.DefaultParams)

	if err != nil {
		common.Fatal(err)
	}

	fmt.Println("Initial password:", initialPw)

	pool, err := pgxpool.New(common.Ctx, "postgres:///xavage")

	if err != nil {
		panic(err)
	}

	var id string
	err = pool.QueryRow(state.Context, "INSERT INTO users (username, password, root, token) VALUES ($1, $2, $3, $4) RETURNING id", name, argon2hash, isAdmin, token).Scan(&id)

	if err != nil {
		common.Fatal(err)
	}

	fmt.Println("User created successfully with ID:", id)
}
