package cmd_utils

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"regexp"
	"time"
)

// KillTimeout timeout for kill signal when exiting a Cmd
var KillTimeout = 1000 * time.Millisecond

// InterruptTimeout timeout for interrupt signal when exiting a Cmd
var InterruptTimeout = 200 * time.Millisecond

// Cmd wraps an exec/Cmd and provides a pipe based interface
type Cmd struct {
	*exec.Cmd

	// Prefix prepended to outputs if provided
	OutputPrefix string
	// ShowOutput prints output to log
	ShowOutput bool
	// DropEmptyLines stops empty lines being received
	DropEmptyLines bool

	// InputChan is the channel attached to the command stdin
	InputChan chan string
	// OutputChan is the channel attached to the command stdout
	OutputChan chan string
}

// Command Creates a command
func Command(name string, arg ...string) *Cmd {
	c := new(Cmd)

	c.OutputPrefix = ""
	c.ShowOutput = false
	c.DropEmptyLines = true

	c.InputChan = nil
	c.OutputChan = nil

	c.Cmd = exec.Command(name, arg...)

	return c
}

// Start wraps Cmd.Start and hooks channels if provided
func (cmd *Cmd) Start() error {

	// Bind output routines if channel exists
	if cmd.OutputChan != nil {
		stdout, err := cmd.Cmd.StdoutPipe()
		if err != nil {
			return err
		}
		go cmd.readCloserToChannel(stdout, cmd.OutputChan)
		stderr, err := cmd.Cmd.StderrPipe()
		if err != nil {
			return err
		}
		go cmd.readCloserToChannel(stderr, cmd.OutputChan)
	}

	// Bind input routine if channel exists
	if cmd.InputChan != nil {
		stdin, err := cmd.Cmd.StdinPipe()
		if err != nil {
			return err
		}
		go cmd.channelToWriteCloser(cmd.InputChan, stdin)
	}

	return cmd.Cmd.Start()
}

// Interrupt sends an os.Interrupt to the process if running
func (cmd *Cmd) Interrupt() {
	if cmd.Process != nil {
		cmd.Process.Signal(os.Interrupt)
	}
}

// Exit a running command
// This attempts a wait, with timeout based interrupt and kill signals
func (cmd *Cmd) Exit() error {

	// Create exit timers
	interruptTimer := time.AfterFunc(InterruptTimeout, func() {
		cmd.Cmd.Process.Signal(os.Interrupt)
	})
	killTimer := time.AfterFunc(KillTimeout, func() {
		cmd.Cmd.Process.Kill()
	})

	// Wait for exit
	err := cmd.Cmd.Wait()

	interruptTimer.Stop()
	killTimer.Stop()

	return err
}

var re = regexp.MustCompile(`(?m)^[\s]+$`)

// Handle output to channel and/or log
func (cmd *Cmd) output(text string) {
	if cmd.DropEmptyLines && re.MatchString(text) {
		return
	}

	var out string
	if cmd.OutputPrefix != "" {
		out = fmt.Sprintf("[%s] %s", cmd.OutputPrefix, text)
	} else {
		out = text
	}

	if cmd.ShowOutput {
		log.Printf("%s", out)
	}
	if cmd.OutputChan != nil {
		cmd.OutputChan <- out
	}

}

// Bind a readable pipe to an output channel for IPC
func (cmd *Cmd) readCloserToChannel(r io.ReadCloser, c chan string) {
	reader := bufio.NewReader(r)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err != io.EOF {
				log.Printf("Pipe read error: %s", err)
			}
			break
		}
		cmd.output(line)
	}
}

// Bind a writable pipe to an input channel for IPC
func (cmd *Cmd) channelToWriteCloser(c chan string, w io.WriteCloser) {
	for {
		select {
		case line, ok := <-c:
			if !ok {
				w.Close()
				break
			}
			_, err := io.WriteString(w, line)
			if err != nil {
				w.Close()
				break
			}
		}
	}
}
