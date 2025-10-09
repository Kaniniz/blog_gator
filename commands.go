package main

import (
	"errors"
	"fmt"
)

type command struct {
	name string
	arguments []string
}

func handlerLogin(s *state, cmd command) error {
	if len(cmd.arguments) < 1 {
		return errors.New("Must enter a username to login")
	}
	err := s.config.SetUser(cmd.arguments[0])
	if err != nil {
		return err
	}
	fmt.Printf("User %s has been set\n", cmd.arguments[0])
	return nil
}

type commands struct {
	handlers map[string]func(*state, command) error
}

func (c *commands) run(s *state, cmd command) error {
	cmd_name := cmd.name
	_, ok := c.handlers[cmd_name]
	if !ok {
		err := fmt.Errorf("Unknown command: %s", cmd.name)
		return err
	}
	err := c.handlers[cmd_name](s, cmd)
	if err != nil {
		return err
	}

	return nil
}

func (c *commands) register(name string, f func(*state, command) error) error {
	c.handlers[name] = f
	return nil
}