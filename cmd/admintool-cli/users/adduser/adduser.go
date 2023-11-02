package adduser

import (
	"admintool-cli/common"
	"fmt"
	"strings"

	"github.com/alexedwards/argon2id"
	"github.com/infinitybotlist/eureka/crypto"
	"github.com/jackc/pgx/v5"
)

func CreateUser(progname string, args []string) {
	var name string

	for _, arg := range args {
		if strings.HasPrefix(arg, "-") {
			continue
		}

		name = arg
		break
	}

	if name == "" {
		common.Fatal("No name provided")
	}

	fmt.Println("Got name:", name)

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

	conn, err := pgx.Connect(common.Ctx, "postgres:///xavage")

	if err != nil {
		panic(err)
	}

	var id string
	err = conn.QueryRow(common.Ctx, "INSERT INTO users (username, password, token) VALUES ($1, $2, $3) RETURNING id", name, argon2hash, token).Scan(&id)

	if err != nil {
		common.Fatal(err)
	}

	fmt.Println("User created successfully with ID:", id)
}
