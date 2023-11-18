//go:build !android

package terminal

var (
	envs   = []string{}
	shells = []string{"bash"}
)
