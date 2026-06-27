package console

import "testing"

func TestCompleteRecoversFromPanic(t *testing.T) {
	c := New("test")
	menu := c.activeMenu()
	menu.Command = nil

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("complete panicked: %v", r)
		}
	}()

	_ = c.complete(nil, 0)

	if menu.Command == nil {
		t.Fatal("complete did not restore the command tree after panic")
	}
}
