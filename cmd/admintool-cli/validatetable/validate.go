package validatetable

import (
	"admintool-cli/common"
	"fmt"
	"os"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/exp/slices"
)

var (
	_pool *pgxpool.Pool
)

func ValidateTable(progname string, args []string) {
	dbName := args[0]
	target := args[1]
	backer := args[2]

	tgtSplit := strings.Split(target, "/")

	if len(tgtSplit) != 2 {
		fmt.Println("invalid target, not in format <target/ref_column>")
		os.Exit(1)
	}

	backSplit := strings.Split(backer, "/")

	if len(backSplit) != 2 {
		fmt.Println("invalid backer, not in format <backer/column>")
		os.Exit(1)
	}

	var err error
	_pool, err = pgxpool.New(common.Ctx, "postgres:///"+dbName)

	if err != nil {
		panic(err)
	}

	sp := common.NewSandboxPool(_pool)

	rows, err := sp.Query(common.Ctx, "SELECT "+tgtSplit[1]+" FROM "+tgtSplit[0]+" WHERE "+tgtSplit[1]+" IS NOT NULL")

	if err != nil {
		panic(err)
	}

	defer rows.Close()

	sp.AllowCommit = true

	delIds := []string{}
	badIds := []string{}

	for rows.Next() {
		var id string

		err := rows.Scan(&id)

		if err != nil {
			panic(err)
		}

		if slices.Contains(delIds, id) {
			fmt.Println("ID", id, "already deleted")
			continue
		}

		// Ensure that the field also exists in the backer table
		var exists bool

		err = sp.QueryRow(common.Ctx, "SELECT EXISTS (SELECT 1 FROM "+backSplit[0]+" WHERE "+backSplit[1]+" = $1)", id).Scan(&exists)

		if err != nil {
			panic(err)
		}

		if !exists {
			fmt.Println("ID", id, "does not exist in", backSplit[0])
			badIds = append(badIds, id)

			var ask bool

			if os.Getenv("S") == "" {
				ask = common.UserInputBoolean("Delete ID " + id + " from " + tgtSplit[0] + "?")
			} else {
				ask = os.Getenv("S") == "y"
			}

			if ask {
				err = sp.Exec(common.Ctx, "DELETE FROM "+tgtSplit[0]+" WHERE "+tgtSplit[1]+" = $1", id)

				if err != nil {
					panic(err)
				}

				fmt.Println("Deleted ID", id, "from", tgtSplit[0])
				delIds = append(delIds, id)
			}
		}
	}

	fmt.Println("Bad IDs:", badIds, "| len:", len(badIds))
}
