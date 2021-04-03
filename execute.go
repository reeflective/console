package gonsole

// execute - The user has entered a command input line, the arguments
// have been processed: we synchronize a few elements of the console,
// then pass these arguments to the command parser for execution and error handling.
func (c *Console) execute(args []string) {

	// Asynchronous messages do not mess with the prompt from now on,
	// until end of execution. Once we are done executing the command,
	// they can again.
	c.isExecuting = true
	defer func() {
		c.isExecuting = false
	}()

	return
}
