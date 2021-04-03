package gonsole

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
