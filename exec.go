package console

import (
	"bytes"
	"fmt"
	"github.com/alexj212/console/parser"
	"github.com/spf13/cobra"
	"os"
)

func (c *Console) exec(cmd *cobra.Command, e *parser.ExecCmd, in *bytes.Buffer, out *bytes.Buffer) error {
	args := append([]string{e.Cmd}, e.Args...)
	args, err := c.runLineHooks(args)
	if err != nil {
		fmt.Printf("executeLine runLineHooks error: %s\n", err.Error())
	}

	cmd.SetArgs(args)
	cmd.SetOut(out)
	cmd.SetErr(out)

	if in != nil {
		cmd.SetIn(in)
	}
	return cmd.Execute()
}

func (c *Console) executeExecCmd(rootCmd *cobra.Command, cmd *parser.ExecCmd) (string, error) {
	out := &bytes.Buffer{}
	var in *bytes.Buffer
	fmt.Printf("executeExecCmd command line %v\n", cmd.Args)

	if err := c.exec(rootCmd, cmd, in, out); err != nil {
		return out.String(), fmt.Errorf("failed to execute: `%s` error: %w", cmd.Args, err)
	}

	// If the command is part of a pipe, use the input from the previous command
	if cmd.Pipe != nil {
		fmt.Printf("executeExecCmd pipe command: %v\n", cmd.Pipe)
		filtered := &bytes.Buffer{}
		line := append([]string{cmd.Pipe.Cmd}, cmd.Pipe.Args...)
		fmt.Printf("executeExecCmd pipe line %v\n", line)
		if err := c.exec(rootCmd, cmd.Pipe, out, filtered); err != nil {
			return out.String(), fmt.Errorf("failed to execute: `%s` error: %w", line, err)
		}

		fmt.Printf("executeExecCmd pipe output: %v\n", filtered.String())
		return filtered.String(), nil
	}

	fmt.Printf("executeExecCmd output: %v\n", out.String())
	return out.String(), nil
}

func (c *Console) ExecuteCommand(rootCmd *cobra.Command, line string) (string, error) {
	if line == "" {
		return "", nil
	}
	outFile, commands, err := parser.ParseCommands(line)
	if err != nil {
		return "", fmt.Errorf("executeLine parsing line `%s` error: %s\n", line, err.Error())
	}
	fmt.Printf("Execute outFile: %v\n", line)
	fmt.Printf("Execute outFile: %v\n", outFile)
	fmt.Printf("Execute commands: %v\n", commands)
	if len(commands) == 0 {
		return "", nil
	}

	var outputBuffer bytes.Buffer
	for _, command := range commands {

		output, err := c.executeExecCmd(rootCmd, command)
		if err != nil {
			fmt.Printf("executeLine %v\n", err)
			break
		}
		outputBuffer.WriteString(output)
	}

	if outFile != "" {
		if err := os.WriteFile(outFile, outputBuffer.Bytes(), 0644); err != nil {
			outputBuffer.WriteString(fmt.Sprintf("executeLine failed to write to file `%v` error: %v\n", outFile, err))
		}
	}
	fmt.Printf("Execute outputBuffer: %v\n", outputBuffer.String())
	return outputBuffer.String(), nil
}
