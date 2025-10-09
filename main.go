package main

import (
	"fmt"
	"os"

	"github.com/Kaniniz/blog_gator/internal/config"
)

func main() {
	response, err := config.Read()
	if err != nil {
		fmt.Println(err)
	}
	cfg := state{
		config: &response,
	}

	cmds := commands{
		handlers: make(map[string]func(*state, command)error),
	}

	cmds.register("login", handlerLogin)
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
	config *config.Config
}