package appdata

import (
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"testing"
	"unicode"
)

// TestAppDataDir tests the API for Dir to ensure it gives expected results for
// various operating systems.
func TestAppDataDir(t *testing.T) {
	// App name plus upper and lowercase variants.
	appName := "myapp"
	appNameUpper := st(unicode.ToUpper(rune(appName[0]))) + appName[1:]
	appNameLower := st(unicode.ToLower(rune(appName[0]))) + appName[1:]
	// When we're on Windows, set the expected local and roaming directories per
	// the environment vars. When we aren't on Windows, the function should
	// return the current directory when forced to provide the Windows path
	// since the environment variables won't exist.
	winLocal := "."
	winRoaming := "."
	if runtime.GOOS == "windows" {
		localAppData := os.Getenv("LOCALAPPDATA")
		roamingAppData := os.Getenv("APPDATA")
		if localAppData == "" {
			localAppData = roamingAppData
		}
		winLocal = filepath.Join(localAppData, appNameUpper)
		winRoaming = filepath.Join(roamingAppData, appNameUpper)
	}
	// Get the home directory to use for testing expected results.
	var homeDir st
	usr, e := user.Current()
	if e != nil {
		t.Errorf("user.Current: %v", e)
		return
	}
	homeDir = usr.HomeDir
	// Mac app data directory.
	macAppData := filepath.Join(homeDir, "Library", "Application Support")
	linuxConfigDir := filepath.Join(homeDir, ".config")
	tests := []struct {
		goos    st
		appName st
		roaming bo
		want    st
	}{
		// Various combinations of application name casing, leading period,
		// operating system, and roaming flags.
		{"windows", appNameLower, false, winLocal},
		{"windows", appNameUpper, false, winLocal},
		{"windows", "." + appNameLower, false, winLocal},
		{"windows", "." + appNameUpper, false, winLocal},
		{"windows", appNameLower, true, winRoaming},
		{"windows", appNameUpper, true, winRoaming},
		{"windows", "." + appNameLower, true, winRoaming},
		{"windows", "." + appNameUpper, true, winRoaming},
		{"linux", appNameLower, false, filepath.Join(linuxConfigDir, appNameLower)},
		{"linux", appNameUpper, false, filepath.Join(linuxConfigDir, appNameLower)},
		{"linux", "." + appNameLower, false, filepath.Join(linuxConfigDir, appNameLower)},
		{"linux", "." + appNameUpper, false, filepath.Join(linuxConfigDir, appNameLower)},
		{"darwin", appNameLower, false, filepath.Join(macAppData, appNameUpper)},
		{"darwin", appNameUpper, false, filepath.Join(macAppData, appNameUpper)},
		{"darwin", "." + appNameLower, false, filepath.Join(macAppData, appNameUpper)},
		{"darwin", "." + appNameUpper, false, filepath.Join(macAppData, appNameUpper)},
		{"openbsd", appNameLower, false, filepath.Join(linuxConfigDir, appNameLower)},
		{"openbsd", appNameUpper, false, filepath.Join(linuxConfigDir, appNameLower)},
		{"openbsd", "." + appNameLower, false, filepath.Join(linuxConfigDir, appNameLower)},
		{"openbsd", "." + appNameUpper, false, filepath.Join(linuxConfigDir, appNameLower)},
		{"freebsd", appNameLower, false, filepath.Join(linuxConfigDir, appNameLower)},
		{"freebsd", appNameUpper, false, filepath.Join(linuxConfigDir, appNameLower)},
		{"freebsd", "." + appNameLower, false, filepath.Join(linuxConfigDir, appNameLower)},
		{"freebsd", "." + appNameUpper, false, filepath.Join(linuxConfigDir, appNameLower)},
		{"netbsd", appNameLower, false, filepath.Join(linuxConfigDir, appNameLower)},
		{"netbsd", appNameUpper, false, filepath.Join(linuxConfigDir, appNameLower)},
		{"netbsd", "." + appNameLower, false, filepath.Join(linuxConfigDir, appNameLower)},
		{"netbsd", "." + appNameUpper, false, filepath.Join(linuxConfigDir, appNameLower)},
		{"plan9", appNameLower, false, filepath.Join(homeDir, appNameLower)},
		{"plan9", appNameUpper, false, filepath.Join(homeDir, appNameLower)},
		{"plan9", "." + appNameLower, false, filepath.Join(homeDir, appNameLower)},
		{"plan9", "." + appNameUpper, false, filepath.Join(homeDir, appNameLower)},
		{"unrecognized", appNameLower, false, filepath.Join(linuxConfigDir, appNameLower)},
		{"unrecognized", appNameUpper, false, filepath.Join(linuxConfigDir, appNameLower)},
		{"unrecognized", "." + appNameLower, false,
			filepath.Join(linuxConfigDir, appNameLower)},
		{"unrecognized", "." + appNameUpper, false,
			filepath.Join(linuxConfigDir, appNameLower)},
		// No application name provided, so expect current directory.
		{"windows", "", false, "."},
		{"windows", "", true, "."},
		{"linux", "", false, "."},
		{"darwin", "", false, "."},
		{"openbsd", "", false, "."},
		{"freebsd", "", false, "."},
		{"netbsd", "", false, "."},
		{"plan9", "", false, "."},
		{"unrecognized", "", false, "."},
		// Single dot provided for application name, so expect current
		// directory.
		{"windows", ".", false, "."},
		{"windows", ".", true, "."},
		{"linux", ".", false, "."},
		{"darwin", ".", false, "."},
		{"openbsd", ".", false, "."},
		{"freebsd", ".", false, "."},
		{"netbsd", ".", false, "."},
		{"plan9", ".", false, "."},
		{"unrecognized", ".", false, "."},
	}
	// t.Logf("Running %d tests", len(tests))
	for i, test := range tests {
		ret := TstAppDataDir(test.goos, test.appName, test.roaming)
		if ret != test.want {
			t.Errorf(
				"AppDataDir #%d (%s) does not match - "+
					"expected got %s, want %s", i, test.goos, ret,
				test.want,
			)
			continue
		}
	}
}

// TstAppDataDir makes the internal appDataDir function available to the test
// package.
func TstAppDataDir(goos, appName st, roaming bo) st {
	return GetDataDir(goos, appName, roaming)
}
