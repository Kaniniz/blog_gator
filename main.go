package main

import _ "github.com/lib/pq"

import (
	"fmt"
	"os"
	"database/sql"

	"github.com/Kaniniz/blog_gator/internal/config"
	"github.com/Kaniniz/blog_gator/internal/database"
)

func main() {
	response, err := config.Read()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	db, err := sql.Open("postgres", "postgres://postgres:gator@localhost:5432/gator")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	dbQueries := database.New(db)

	cfg := state{
		config: &response,
		db: dbQueries,
	}

	cmds := commands{
		handlers: make(map[string]func(*state, command)error),
	}

	cmds.register("login", handlerLogin)
	cmds.register("register", handlerRegister)
	cmds.register("reset", handlerResetUsers)
	cmds.register("users", handlerGetUsers)
	cmds.register("agg", handlerAgg)

	if len(os.Args) < 2 {
		fmt.Println("ERROR: No command name")
		os.Exit(1)
	}
	//Read the user input and create the command struct.
	cmd := command{
		name: os.Args[1],
		arguments: os.Args[2:],
	}

	err = cmds.run(&cfg, cmd)
	if err != nil {
		fmt.Println("Error during command execution:", err)
		os.Exit(1)
	}

	os.Exit(0)
}

type state struct {
	db  *database.Queries
	config *config.Config
}