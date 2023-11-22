module github.com/pygrum/console

go 1.21

require (
	github.com/kballard/go-shellquote v0.0.0-20180428030007-95032a82bc51
	github.com/reeflective/readline v1.0.9
	github.com/rsteube/carapace v0.43.3
	github.com/spf13/cobra v1.7.0
	github.com/spf13/pflag v1.0.5
	golang.org/x/exp v0.0.0-20220909182711-5c715a9e8561
)

require (
	github.com/google/go-cmp v0.5.9 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/kr/pretty v0.3.1 // indirect
	github.com/rivo/uniseg v0.4.4 // indirect
	golang.org/x/sys v0.8.0 // indirect
	golang.org/x/term v0.8.0 // indirect
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/rsteube/carapace v0.43.3 => github.com/reeflective/carapace v0.25.2-0.20230816093630-a30f5184fa0d
