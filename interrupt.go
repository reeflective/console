package console

// AddInterrupt registers a handler to run when the console receives a given
// interrupt error from the underlying readline shell. Mainly two interrupt
// signals are concerned: io.EOF (returned when pressing CtrlD), and console.ErrCtrlC.
// Many will want to use this to switch menus. Note that these interrupt errors only
// work when the console is NOT currently executing a command, only when reading input.
func (m *Menu) AddInterrupt(err error, handler func(c *Console)) {
	m.mutex.RLock()
	m.interruptHandlers[err] = handler
	m.mutex.RUnlock()
}

// DelInterrupt removes one or more interrupt handlers from the menu registered ones.
// If no error is passed as argument, all handlers are removed.
func (m *Menu) DelInterrupt(errs ...error) {
	m.mutex.RLock()
	if len(errs) == 0 {
		m.interruptHandlers = make(map[error]func(c *Console))
	} else {
		for _, err := range errs {
			delete(m.interruptHandlers, err)
		}
	}
	m.mutex.RUnlock()
}

func (m *Menu) handleInterrupt(err error) {
	if handler := m.interruptHandlers[err]; handler != nil {
		handler(m.console)
	}
}
