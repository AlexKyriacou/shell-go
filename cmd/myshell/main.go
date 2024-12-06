package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type Command func([]string)

var builtins = make(map[string]Command)

func main() {
	builtins = map[string]Command{
		"echo": echo,
		"exit": exit,
		"type": typeCommand,
	}
	for {
		fmt.Fprint(os.Stdout, "$ ")

		// Wait for user input
		commandRaw, err := bufio.NewReader(os.Stdin).ReadString('\n')
		if err != nil {
			fmt.Println("Error In User Input")
		}
		command := strings.Split(strings.TrimSpace(commandRaw), " ")
		if commandHandler, exists := builtins[command[0]]; exists {
			commandHandler(command[1:])
		} else {
			fmt.Println(command[0] + ": command not found")
		}
	}
}

func typeCommand(args []string) {
	if len(args) > 1 {
		fmt.Println("Invalid number of arguements for command type. Expected 1, Got", len(args))
	}
	if _, exists := builtins[args[0]]; exists {
		fmt.Println(args[0] + " is a shell builtin")
	} else if path, err := findExecutablePath(args[0]); err == nil {
		fmt.Println(args[0], "is", path)
	} else {
		fmt.Println(args[0] + ": not found")
	}
}

func findExecutablePath(target string) (string, error) {
	pathEnv, isSet := os.LookupEnv("PATH")
	if !isSet {
		return "", errors.New("PATH variable is not set")
	}
	paths := strings.Split(pathEnv, ":")
	for _, dir := range paths {
		entries, _ := os.ReadDir(dir)
		for _, entry := range entries {
			if entry.Name() == target {
				return filepath.Abs(filepath.Join(dir, entry.Name()))
			}
		}
	}
	return "", errors.New("executable file not found in PATH")
}

func echo(args []string) {
	fmt.Println(strings.Join(args, " "))
}

func exit(args []string) {
	if len(args) != 1 {
		fmt.Println("Invalid number of arguements for command exit. Expected 1, Got", len(args))
	} else {
		exitCode, err := strconv.Atoi(args[0])
		if err != nil {
			fmt.Println("Invalid exit code " + args[0])
		}
		os.Exit(exitCode)
	}
}
