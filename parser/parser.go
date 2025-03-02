package parser

import (
	"errors"
	"strings"
)

type ExecCmd struct {
	Cmd  string
	Args []string
	Pipe *ExecCmd
	Line int // Line number for error reporting
}

func (c *ExecCmd) String() string {
	if c.Pipe != nil {
		return c.Cmd + " " + strings.Join(c.Args, " ") + " | " + c.Pipe.String()
	}
	return c.Cmd + " " + strings.Join(c.Args, " ")
}

// ParseCommands processes multi-line input into executable commands
func ParseCommands(input string) (string, []*ExecCmd, error) {
	var outputFile string
	var commands []*ExecCmd

	// Remove comments and handle multi-line commands
	lines := strings.Split(input, "\n")
	filteredLines := []string{}
	var currentLine string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "#") {
			continue
		}
		if strings.HasSuffix(line, "\\") {
			currentLine += strings.TrimSuffix(line, "\\") + " "
			continue
		}
		currentLine += line
		filteredLines = append(filteredLines, currentLine)
		currentLine = ""
	}
	input = strings.Join(filteredLines, " ")

	// Check for output redirection first
	if strings.Contains(input, ">") {
		parts := strings.Split(input, ">")
		if len(parts) != 2 {
			return "", nil, errors.New("invalid output redirection syntax")
		}
		input = strings.TrimSpace(parts[0])
		if outputFile != "" {
			return "", nil, errors.New("multiple output redirections are not allowed")
		}
		outputFile = strings.TrimSpace(parts[1])
	}

	// Split the input by ';' to handle multiple commands
	commandGroups := strings.Split(input, ";")

	for _, group := range commandGroups {
		group = strings.TrimSpace(group)
		if group == "" {
			continue
		}

		// Parse piped commands
		var prevCmd *ExecCmd
		pipeParts := strings.Split(group, "|")
		for _, part := range pipeParts {
			part = strings.TrimSpace(part)
			if part == "" {
				continue
			}

			cmdParts := strings.Fields(part)
			if len(cmdParts) == 0 {
				return "", nil, errors.New("invalid command syntax")
			}

			cmd := &ExecCmd{
				Cmd:  cmdParts[0],
				Args: cmdParts[1:],
			}

			if prevCmd != nil {
				prevCmd.Pipe = cmd
			} else {
				commands = append(commands, cmd)
			}

			prevCmd = cmd
		}
	}

	return outputFile, commands, nil
}
