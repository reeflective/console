package args

import (
	"github.com/rsteube/carapace"
	"github.com/rsteube/carapace/pkg/style"
)

// IP is another argument field, but which implements
// a slightly more complicated completion interface.
type IP []string

// Complete produces completions for the IP type.
func (ip *IP) Complete(ctx carapace.Context) carapace.Action {
	action := carapace.ActionStyledValuesDescribed(
		"23:23:234:34ef:343f:47ca", "An IPv6 address", style.BrightGreen,
		"::1", "a test address", style.BrightGreen,
		"10.10.10.10", "An intruder", style.Blue,
	).Tag("IPv6 addresses").Invoke(ctx).Filter(ctx.Args).ToA()

	return action
}

// Host is another type used as a positional argument.
type Host string

// Complete generates completions for the Host type.
func (p *Host) Complete(ctx carapace.Context) carapace.Action {
	action := carapace.ActionStyledValuesDescribed(
		"192.168.1.1", "A first ip address", style.BgBlue,
		"192.168.3.12", "a second address", style.BrightGreen,
		"10.203.23.45", "and a third one", style.BrightCyan,
		"127.0.0.1", "and a third one", style.BrightCyan,
		"219.293.91.10", "", style.Blue,
	).Tag("IPv4 addresses").Invoke(ctx).Filter(ctx.Args).ToA()

	return action
}

// Proxy is another type used as a positional argument.
type Proxy string

// Complete generates completions for the Proxy type.
func (p *Proxy) Complete(ctx carapace.Context) carapace.Action {
	action := carapace.ActionValuesDescribed(
		"github.com", "A first ip address",
		"google.com", "a second address",
		"blue-team.com", "and a third one",
	).Tag("host domains").Invoke(ctx).Filter(ctx.Args).ToA()

	return action
}
