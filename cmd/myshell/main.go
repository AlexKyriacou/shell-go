package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"unicode"
)

type Command func([]string)

var builtins = make(map[string]Command)

func main() {
	builtins = map[string]Command{
		"echo": echo,
		"exit": exit,
		"type": typeCommand,
		"pwd":  pwd,
		"cd":   cd,
	}

	for {
		fmt.Fprint(os.Stdout, "$ ")

		commandRaw, err := bufio.NewReader(os.Stdin).ReadString('\n')
		if err != nil {
			fmt.Println("Error In User Input")
		}

		command := parseRawCommand(commandRaw)
		if commandHandler, exists := builtins[command[0]]; exists {
			commandHandler(command[1:])
		} else if path, err := findExecutablePath(command[0]); err == nil {
			cmd := exec.Command(path, command[1:]...)
			cmd.Env = os.Environ()
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			err := cmd.Run()
			if err != nil {
				fmt.Println(err)
			}
		} else {
			fmt.Println(command[0] + ": command not found")
		}
	}
}

func parseRawCommand(command string) []string {
	var args []string
	var arg strings.Builder
	inSingleQuotes := false
	inDoubleQuotes := false
	escaped := false

	for _, char := range command {
		switch {
		case char == '\'':
			if escaped && inDoubleQuotes {
				arg.WriteRune('\\')
			}
			if inDoubleQuotes || escaped {
				arg.WriteRune(char)
			} else {
				inSingleQuotes = !inSingleQuotes
			}
			escaped = false
		case char == '"':
			if inSingleQuotes || escaped {
				arg.WriteRune(char)
			} else {
				inDoubleQuotes = !inDoubleQuotes
			}
			escaped = false
		case char == '\\':
			if inSingleQuotes || escaped {
				arg.WriteRune(char)
				escaped = false
			} else {
				escaped = true
			}
		case unicode.IsSpace(char):
			if escaped && (inDoubleQuotes || inSingleQuotes){
				arg.WriteRune('\\')
			}
			if inSingleQuotes || inDoubleQuotes || escaped {
				arg.WriteRune(char)
			} else if arg.Len() > 0 {
				args = append(args, arg.String())
				arg.Reset()
			}
			escaped = false
		default:
			if escaped && inDoubleQuotes {
				arg.WriteRune('\\')
			}
			arg.WriteRune(char)
			escaped = false
		}
	}

	if arg.Len() > 0 {
		args = append(args, arg.String())
	}

	return args
}

func cd(args []string) {
	if len(args) != 1 {
		fmt.Println("Invalid number of arguements for command cd. Expected 1, Got", len(args))
		return
	}

	var targetDir string
	if args[0] == "~" {
		homeDir, isSet := os.LookupEnv("HOME")
		if !isSet {
			fmt.Println("HOME variable is not set")
			return
		}
		targetDir = homeDir
	} else {
		targetDir = args[0]
	}

	err := os.Chdir(targetDir)
	if err != nil {
		fmt.Println("cd:", targetDir+": No such file or directory")
		return
	}
}

func pwd(_ []string) {
	dirPath, err := os.Getwd()
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(dirPath)
}

func typeCommand(args []string) {
	if len(args) != 1 {
		fmt.Println("Invalid number of arguements for command type. Expected 1, Got", len(args))
		return
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

	paths := strings.Split(pathEnv, string(os.PathListSeparator))
	for _, dir := range paths {
		fullPath := filepath.Join(dir, target)
		if _, err := os.Stat(fullPath); err == nil {
			return fullPath, nil
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
		return
	}

	exitCode, err := strconv.Atoi(args[0])
	if err != nil {
		fmt.Println("Invalid exit code " + args[0])
	}
	
	os.Exit(exitCode)
}
