package createmultipleusers

import (
	"admintool-cli/common"
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/alexedwards/argon2id"
	"github.com/infinitybotlist/eureka/crypto"
	"github.com/jackc/pgx/v5/pgxpool"
)

func xkcdpass() (string, error) {
	cmd, err := exec.Command("xkcdpass", "--min", "5", "--max", "8", "-n", "4").Output()

	if err != nil {
		return "", err
	}

	cmdStr := strings.TrimSpace(strings.ReplaceAll(string(cmd), "\n", ""))

	return cmdStr, nil
}

func CreateMultipleUsers(progname string, args []string) {
	var usernames []string

	var buffer = bufio.NewReader(os.Stdin)
	for {
		// Get input
		var input string
		input, err := buffer.ReadString('\n')

		if err != nil {
			common.Fatal(err)
		}

		input = strings.TrimSpace(input)

		if os.Getenv("STRIP_SPECIFIC_CHARS") != "" {
			charsToStrip := strings.Split(os.Getenv("STRIP_SPECIFIC_CHARS"), "")

			for _, char := range charsToStrip {
				input = strings.ReplaceAll(input, char, "")
			}
		}

		if input == "" {
			continue
		}

		if input == "EOF" {
			break
		}

		usernames = append(usernames, input)
	}

	common.StatusBoldYellow("Creating the below users:\n\n")

	for _, username := range usernames {
		common.StatusBoldYellow(username)
	}

	common.StatusBoldYellow(len(usernames), "users to create\n\n")

	inp := common.UserInputBoolean("Continue?")

	if !inp {
		common.Fatal("Aborted")
	}

	pool, err := pgxpool.New(common.Ctx, "postgres:///xavage")

	if err != nil {
		common.Fatal(err)
	}

	tx, err := pool.Begin(common.Ctx)

	if err != nil {
		common.Fatal(err)
	}

	defer tx.Rollback(common.Ctx)

	for _, username := range usernames {
		pass, err := xkcdpass()

		if err != nil {
			common.Fatal(err)
			return
		}

		token := crypto.RandString(512)

		argon2hash, err := argon2id.CreateHash(pass, argon2id.DefaultParams)

		if err != nil {
			common.Fatal(err)
		}

		if err != nil {
			panic(err)
		}

		var id string
		err = tx.QueryRow(common.Ctx, "INSERT INTO users (username, password, token) VALUES ($1, $2, $3) RETURNING id", username, argon2hash, token).Scan(&id)

		if err != nil {
			common.Fatal(err)
		}

		fmt.Printf("[%s] %s: %s\n", id, username, pass)
	}

	err = tx.Commit(common.Ctx)

	if err != nil {
		common.Fatal(err)
	}
}
