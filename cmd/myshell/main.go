package main

import (
	"errors"
	"fmt"
	"golang.org/x/term"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"unicode"
)

type CommandHandler func([]string)

type Command []string

var builtins map[string]CommandHandler

func init() {
	builtins = map[string]CommandHandler{
		"echo": echo,
		"exit": exit,
		"type": typeCommand,
		"pwd":  pwd,
		"cd":   cd,
	}
}

func main() {
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		panic(err)
	}
	defer term.Restore(int(os.Stdin.Fd()), oldState)

	for {
		fmt.Fprint(os.Stdout, "$ ")

		var commandRaw []byte
		buf := make([]byte, 1)
	readLoop:
		for {
			os.Stdin.Read(buf)
			switch buf[0] {
			case 3: // Ctrl+C
				return
			case 127, 8: // Backspace (127) or Ctrl+H (8)
				if len(commandRaw) > 0 {
					// Remove last character from input
					commandRaw = commandRaw[:len(commandRaw)-1]
					// Move cursor back, print space, move cursor back again
					fmt.Print("\b \b")
				}
			case 13: // Enter
				fmt.Println()
				break readLoop
			case 9: // Tab
				autocompleteMatches := autoCompleteCommand(commandRaw)
				commandRaw = []byte(autocompleteMatches[0])
				fmt.Printf("\r%*s\r", len(commandRaw), "")
				fmt.Print("$ " + string(commandRaw) + " ")
				continue
			default:
				if buf[0] >= 32 && buf[0] <= 126 { // printable characters
					commandRaw = append(commandRaw, buf[0])
					fmt.Print(string(buf[0]))
				}
			}
		}

		var command Command = parseRawCommand(string(commandRaw))
		originalStdout := os.Stdout
		originalStErr := os.Stderr
		if command.hasInputRedirection() {
			command.redirectInput()
			command = command[:len(command)-2]
		}
		if commandHandler, exists := builtins[command[0]]; exists {
			commandHandler(command[1:])
		} else if path, err := findExecutablePath(command[0]); err == nil {
			cmd := exec.Command(path, command[1:]...)
			cmd.Env = os.Environ()
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			cmd.Run()
		} else {
			fmt.Println(command[0] + ": command not found")
		}
		os.Stdout = originalStdout
		os.Stderr = originalStErr
	}
}

func autoCompleteCommand(commandRaw []byte) []string {
	var matches []string
	for builtinName := range builtins {
		if strings.HasPrefix(builtinName, string(commandRaw)) {
			matches = append(matches, builtinName)
		}
	}
	return matches
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
			if escaped && (inDoubleQuotes || inSingleQuotes) {
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

func (command Command) hasInputRedirection() bool {
	inputRedirectionOperators := []string{
		">",
		"1>",
		"2>",
		">>",
		"1>>",
		"2>>",
	}
	if len(command) < 2 {
		return false
	}
	inputRedirectionArguement := command[len(command)-2]
	if len(inputRedirectionArguement) < 1 {
		return false
	}
	for _, operator := range inputRedirectionOperators {
		if operator == inputRedirectionArguement {
			return true
		}
	}
	return false
}

func (command Command) redirectInput() error {
	if len(command) < 2 {
		return nil
	}

	redirectOp := command[len(command)-2]
	targetPath := command[len(command)-1]
	ensureDir(targetPath)

	// Determine file flags and target based on operation
	var (
		flags  int
		target **os.File
	)

	// Set flags based on whether it's append mode
	if strings.HasSuffix(redirectOp, ">>") {
		flags = os.O_WRONLY | os.O_CREATE | os.O_APPEND
	} else {
		flags = os.O_WRONLY | os.O_CREATE | os.O_TRUNC
	}

	// Set target based on whether it's stderr
	if strings.HasPrefix(redirectOp, "2") {
		target = &os.Stderr
	} else {
		target = &os.Stdout
	}

	// Open file and set target
	file, err := os.OpenFile(targetPath, flags, 0644)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return err
	}
	*target = file

	// Ensure the file is closed at the end of the command execution
	defer func() {
		file.Close()
	}()

	// Even if there's no output, the file is now created or updated
	return nil
}

func ensureDir(fileName string) {
	dirName := filepath.Dir(fileName)
	if _, serr := os.Stat(dirName); serr != nil {
		merr := os.MkdirAll(dirName, os.ModePerm)
		if merr != nil {
			panic(merr)
		}
	}
}
