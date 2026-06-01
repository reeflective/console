package console

// AddInterrupt registers a handler to run when the console receives
// a given interrupt error from the underlying readline shell.
//
// On most systems, the following errors will be returned with keypresses:
// - Linux/MacOS/Windows : Ctrl-C will return os.Interrupt.
//
// Many will want to use this to switch menus. Note that these interrupt errors only
// work when the console is NOT currently executing a command, only when reading input.
func (m *Menu) AddInterrupt(err error, handler func(c *Console)) {
	m.mutex.Lock()
	m.interruptHandlers[err] = handler
	m.mutex.Unlock()
}

// DelInterrupt removes one or more interrupt handlers from the menu registered ones.
// If no error is passed as argument, all handlers are removed.
func (m *Menu) DelInterrupt(errs ...error) {
	m.mutex.Lock()
	if len(errs) == 0 {
		m.interruptHandlers = make(map[error]func(c *Console))
	} else {
		for _, err := range errs {
			delete(m.interruptHandlers, err)
		}
	}
	m.mutex.Unlock()
}

func (m *Menu) handleInterrupt(err error) {
	m.console.isExecuting.Store(true)
	defer m.console.isExecuting.Store(false)

	// TODO: this is not a very, very safe way of comparing
	// errors. I'm not sure what to right now with this, but
	// from my (unreliable) expectations and usage, I see and
	// use things like errors.New(os.Interrupt.String()), so
	// the string itself is likely to change in the future.
	//
	// But if people use their own third-party errors... nothing is guaranteed.
	//
	// Snapshot the matching handlers under the lock, then run them once
	// released: a handler is free to mutate the menu (e.g. SwitchMenu)
	// without deadlocking, and the map can't be written mid-iteration.
	m.mutex.RLock()
	matched := make([]func(c *Console), 0, len(m.interruptHandlers))
	for herr, handler := range m.interruptHandlers {
		if err.Error() == herr.Error() {
			matched = append(matched, handler)
		}
	}
	m.mutex.RUnlock()

	for _, handler := range matched {
		handler(m.console)
	}
}
