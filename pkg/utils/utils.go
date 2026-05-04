// Package utils is a utility package that helps with the application
package utils

import (
	"context"
	"os/exec"
	"runtime"
)

// OpenURL opens a URL in the system default browser.
func OpenURL(url string) {
	ctx := context.Background()
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.CommandContext(ctx, "open", url)
	case "linux":
		cmd = exec.CommandContext(ctx, "xdg-open", url)
	default:
		cmd = exec.CommandContext(ctx, "rundll32", "url.dll,FileProtocolHandler", url)
	}
	_ = cmd.Start()
}
