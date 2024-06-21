//go:build brew

package config

func defaultHome() string {
	// this value is replaced by homebrew with the user's share location
	return "/usr/share/siegfried"
}
