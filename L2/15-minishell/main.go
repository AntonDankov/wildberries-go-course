package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"strings"

	"golang.org/x/sys/windows"
	"golang.org/x/term"
)

var currentCmd *exec.Cmd

func main() {
	signal.Ignore(os.Interrupt)

	inputChan := make(chan string, 1)
	doneChan := make(chan struct{})
	go func() {
		oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
		if err != nil {
			panic(err)
		}
		defer term.Restore(int(os.Stdin.Fd()), oldState)

		fmt.Print("mini-shell>")

		var line []byte
		buf := make([]byte, 1)

		for {
			select {
			default:
				n, err := os.Stdin.Read(buf)
				if err != nil {
					fmt.Println("\nFailed to read input")
					return
				}

				if n == 0 {
					continue
				}

				switch buf[0] {
				case 0x04: // Ctrl+D
					if currentCmd != nil && currentCmd.Process != nil {
						killProcessTree(currentCmd.Process.Pid)
					}
					close(doneChan)
					fmt.Println("\nExiting shell...")
					return

				case 0x03: // Ctrl+C

					if currentCmd != nil && currentCmd.Process != nil {
						killProcessTree(currentCmd.Process.Pid)
					} else {
						fmt.Print("^C\r\n")
						fmt.Print("mini-shell>")
					}
					line = nil

				case '\r', '\n':
					if len(line) > 0 {
						text := string(line)
						fmt.Print("\r\n")
						inputChan <- text
						line = nil
					} else {
						fmt.Print("\nmini-shell>")
					}

				case 0x7F, 0x08: // Backspace
					if len(line) > 0 {
						line = line[:len(line)-1]
						// Remove character from terminal
						fmt.Print("\b \b")
					}

				default:
					if buf[0] >= 32 && buf[0] < 127 && currentCmd == nil {
						line = append(line, buf[0])
						fmt.Printf("%c", buf[0])
					}
				}
			}
		}
	}()

	for {
		select {
		case text, ok := <-inputChan:
			if !ok {
				return
			}

			processInput(text)
			fmt.Print("\nmini-shell>")
		case <-doneChan:
			return
		}
	}
}

func killProcessTree(pid int) error {
	kill := exec.Command("taskkill", "/F", "/T", "/PID", strconv.Itoa(pid))
	return kill.Run()
}

func processInput(input string) {
	commands := strings.Split(input, "|")
	if err := validateCommands(commands); err != nil {
		fmt.Printf("error: %v\n", err)
		return
	}
	switch len(commands) {
	case 0:
		return
	case 1:
		result, err := processCommand(commands[0], "", os.Stdin, os.Stdout)
		if err != nil {
			fmt.Printf("error: %v\n", err)
		} else if len(result) > 0 {
			fmt.Println(result)
		}
	default:
		processPipeline(commands)
	}
}

func validateCommands(commands []string) error {
	for i, command := range commands {
		if len(command) == 0 {
			return fmt.Errorf("commnd %d have size 0", i+1)
		}
	}
	return nil
}

// PipelineCommand holds command with pipeline configuration
type PipelineCommand struct {
	isBuiltin  bool
	cmd        *exec.Cmd
	stdin      io.Reader
	stdout     io.WriteCloser
	pipeWriter *io.PipeWriter
	name       string
	args       []string
}

func processPipeline(commands []string) {
	var pipelineCmds []PipelineCommand
	oldState, _ := term.GetState(int(os.Stdin.Fd()))
	if oldState != nil {
		term.Restore(int(os.Stdin.Fd()), oldState)
	}

	var currentReader io.Reader = os.Stdin

	for i, cmdStr := range commands {
		parts := strings.Fields(strings.TrimSpace(cmdStr))
		var pipelineCmd PipelineCommand

		pipelineCmd.name = parts[0]
		pipelineCmd.args = parts[1:]
		pipelineCmd.stdin = currentReader
		if i == len(commands)-1 {
			pipelineCmd.stdout = os.Stdout
		} else {
			reader, writer := io.Pipe()
			pipelineCmd.stdout = writer
			pipelineCmd.pipeWriter = writer
			currentReader = reader
		}
		if isBuiltin(parts[0]) {
			pipelineCmd.isBuiltin = true
		} else {
			pipelineCmd.isBuiltin = false
			pipelineCmd.cmd = exec.Command(parts[0], parts[1:]...)
			pipelineCmd.cmd.Stdin = pipelineCmd.stdin
			pipelineCmd.cmd.Stdout = pipelineCmd.stdout
		}
		pipelineCmds = append(pipelineCmds, pipelineCmd)
	}

	for i := 0; i < len(pipelineCmds); i++ {
		if !pipelineCmds[i].isBuiltin {
			pipelineCmds[i].cmd.Start()
		} else {
			pc := pipelineCmds[i]
			go func(pipelineCmd PipelineCommand) {
				defer pipelineCmd.stdout.Close()
				for {
					result, _ := processCommand(pipelineCmd.name, strings.Join(pipelineCmd.args, ","), pipelineCmd.stdin, pipelineCmd.stdout)

					pipelineCmd.stdout.Write([]byte(result + "\n"))
					return
				}
			}(pc)
		}
	}

	// Wait for all to finish
	for i, pipelineCmd := range pipelineCmds {
		if !pipelineCmd.isBuiltin {
			currentCmd = pipelineCmd.cmd
			pipelineCmd.cmd.Wait()
			if i != len(pipelineCmds)-1 {
				pipelineCmd.stdout.Close()
			}
		}
	}
	currentCmd = nil
	if oldState != nil {
		term.MakeRaw(int(os.Stdin.Fd()))
	}
}

func isBuiltin(command string) bool {
	builtins := []string{"cd", "pwd", "echo", "ps", "kill", "exit"}
	for _, b := range builtins {
		if command == b {
			return true
		}
	}
	return false
}

func processCommand(input string, pipeLineInput string, stdin io.Reader, stdout io.Writer) (string, error) {
	command, args, _ := strings.Cut(strings.Trim(input, " "), " ")

	result := ""
	var err error
	switch command {
	case "cd":
		_, err = executeCd(args)
	case "kill":

		_, err = executeKill(args)

	case "echo":
		result, err = executeEcho(args + pipeLineInput)
	case "pwd":
		result, err = executePwd()

	case "ps":
		result, err = executePs()
	case "exit":
		os.Exit(0)
	default:
		result, err = executeCommand(command, args, stdin, stdout)
	}

	return result, err
}

func executeCommand(command string, args string, stdin io.Reader, stdout io.Writer) (string, error) {
	cmdArgs := []string{"/C", command}

	if len(args) > 0 {
		cmdArgs = append(cmdArgs, strings.Fields(args)...)
	}

	oldState, _ := term.GetState(int(os.Stdin.Fd()))
	if oldState != nil {
		term.Restore(int(os.Stdin.Fd()), oldState)
	}

	currentCmd = exec.Command("cmd", cmdArgs...)

	currentCmd.Stdin = stdin
	currentCmd.Stdout = stdout

	currentCmd.Start()
	err := currentCmd.Wait()
	if oldState != nil {
		term.MakeRaw(int(os.Stdin.Fd()))
	}
	currentCmd = nil

	return "", err
}

func executeCd(input string) (string, error) {
	return "", os.Chdir(input)
}

func executeKill(input string) (string, error) {
	if len(input) == 0 {
		err := fmt.Errorf("kill: missed PID")
		return " ", err
	}
	pid, err := strconv.Atoi(input)
	if err != nil {
		return "", nil
	}

	process, err := os.FindProcess(pid)
	if err != nil {
		return "", nil
	}
	handle, err := windows.OpenProcess(windows.PROCESS_TERMINATE, false, uint32(process.Pid))
	if err != nil {
		return "", err
	}
	defer windows.CloseHandle(handle)
	err = windows.TerminateProcess(handle, 1)
	return "", err
}

func executeEcho(input string) (string, error) {
	return input, nil
}

func executePwd() (string, error) {
	pwd, err := os.Getwd()
	return pwd, err
}

func executePs() (string, error) {
	cmd := exec.Command("tasklist")
	output, err := cmd.Output()

	return string(output), err
}
