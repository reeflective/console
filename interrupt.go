package console

import "errors"

// AddInterrupt registers a handler to run when the console receives
// a given interrupt error from the underlying readline shell.
//
// On most systems, the following errors will be returned with keypresses:
// - Linux/MacOS/Windows : Ctrl-C will return os.Interrupt.
//
// The incoming error is matched against the registered one with errors.Is
// first (so wrapped errors and sentinel values work as expected), falling
// back to comparing their messages for errors that are merely value-equal
// (e.g. two distinct errors.New with the same text).
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

	// Match with errors.Is first so sentinel and wrapped errors behave
	// correctly, then fall back to comparing messages for errors that are
	// only value-equal (the historically supported errors.New(...) pattern).
	//
	// Snapshot the matching handlers under the lock, then run them once
	// released: a handler is free to mutate the menu (e.g. SwitchMenu)
	// without deadlocking, and the map can't be written mid-iteration.
	m.mutex.RLock()
	matched := make([]func(c *Console), 0, len(m.interruptHandlers))
	for herr, handler := range m.interruptHandlers {
		if errors.Is(err, herr) || err.Error() == herr.Error() {
			matched = append(matched, handler)
		}
	}
	m.mutex.RUnlock()

	for _, handler := range matched {
		handler(m.console)
	}
}
