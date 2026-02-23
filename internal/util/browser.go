package util

import (
	"os"
	"os/exec"
	"runtime"
	"strings"
)

// OpenBrowser opens url using HNCLI_OPEN if set, otherwise the system browser.
//
// HNCLI_OPEN is treated as a sh -c command. Use {} as a placeholder for the
// URL; if absent the URL is appended as the last argument. Examples:
//
//	HNCLI_OPEN="xdg-open"                          # explicit browser (default)
//	HNCLI_OPEN="echo {} | xclip -selection clipboard"  # Linux clipboard
//	HNCLI_OPEN="echo {} | wl-copy"                 # Wayland clipboard
//	HNCLI_OPEN="echo {} | pbcopy"                  # macOS clipboard
func OpenBrowser(url string) error {
	if tmpl := os.Getenv("HNCLI_OPEN"); tmpl != "" {
		var sh string
		if strings.Contains(tmpl, "{}") {
			sh = strings.ReplaceAll(tmpl, "{}", url)
		} else {
			sh = tmpl + " " + url
		}
		return exec.Command("sh", "-c", sh).Start()
	}
	return openDefault(url)
}

// openDefault opens url in the system default browser.
func openDefault(url string) error {
	var cmd string
	var args []string
	switch runtime.GOOS {
	case "darwin":
		cmd = "open"
		args = []string{url}
	case "windows":
		cmd = "rundll32"
		args = []string{"url.dll,FileProtocolHandler", url}
	default:
		cmd = "xdg-open"
		args = []string{url}
	}
	return exec.Command(cmd, args...).Start()
}
