package executor

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNormalizeGradleCacheURL(t *testing.T) {
	tests := []struct {
		in, want string
		wantErr  bool
	}{
		{"", "", false},
		{"http://coord:8080/cache/gradle", "http://coord:8080/cache/gradle/", false},
		{" https://coord/cache/gradle/ ", "https://coord/cache/gradle/", false},
		{"ftp://coord/cache", "", true},
		{"coord:8080/cache", "", true},
		{"http://u:p@coord/cache/", "", true},
		{"http://coord/cache/'+evil+'", "", true},
		{"http://coord/cache/${x}", "", true},
	}
	for _, tt := range tests {
		got, err := NormalizeGradleCacheURL(tt.in)
		if (err != nil) != tt.wantErr || got != tt.want {
			t.Errorf("NormalizeGradleCacheURL(%q) = %q, %v; want %q, err=%v", tt.in, got, err, tt.want, tt.wantErr)
		}
	}
}

func writeAndroidProject(t *testing.T, settingsName string) string {
	t.Helper()
	dir := t.TempDir()
	android := filepath.Join(dir, "android")
	if err := os.MkdirAll(android, 0o755); err != nil {
		t.Fatal(err)
	}
	if settingsName != "" {
		if err := os.WriteFile(filepath.Join(android, settingsName), []byte("include ':app'\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(android, "gradle.properties"), []byte("android.useAndroidX=true"), 0o644); err != nil {
		t.Fatal(err)
	}
	return dir
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func TestConfigureGradleCache(t *testing.T) {
	cache := GradleCache{URL: "http://coord:8080/cache/gradle/", Token: "secret-token"}
	tests := []struct {
		settings string
		want     []string
	}{
		{"settings.gradle", []string{"remote(HttpBuildCache)", "url = 'http://coord:8080/cache/gradle/'", "allowInsecureProtocol = true"}},
		{"settings.gradle.kts", []string{"remote<HttpBuildCache>", `url = uri("http://coord:8080/cache/gradle/")`, "isPush = true"}},
	}
	for _, tt := range tests {
		t.Run(tt.settings, func(t *testing.T) {
			dir := writeAndroidProject(t, tt.settings)
			for i := 0; i < 2; i++ { // second call must not append again
				if err := configureGradleCache(dir, cache); err != nil {
					t.Fatal(err)
				}
			}
			settings := readFile(t, filepath.Join(dir, "android", tt.settings))
			if !strings.HasPrefix(settings, "include ':app'\n") || strings.Count(settings, gradleCacheMarker) != 1 {
				t.Fatalf("settings not appended exactly once:\n%s", settings)
			}
			for _, w := range tt.want {
				if !strings.Contains(settings, w) {
					t.Errorf("settings missing %q:\n%s", w, settings)
				}
			}
			if strings.Contains(settings, cache.Token) {
				t.Fatal("token written into the settings file")
			}
			props := readFile(t, filepath.Join(dir, "android", "gradle.properties"))
			if !strings.Contains(props, "android.useAndroidX=true\norg.gradle.caching=true") {
				t.Fatalf("gradle.properties = %q", props)
			}
		})
	}
}

func TestConfigureGradleCache_NoOp(t *testing.T) {
	// Disabled: nothing changes.
	dir := writeAndroidProject(t, "settings.gradle")
	if err := configureGradleCache(dir, GradleCache{}); err != nil {
		t.Fatal(err)
	}
	if got := readFile(t, filepath.Join(dir, "android", "settings.gradle")); got != "include ':app'\n" {
		t.Fatalf("settings changed while disabled: %q", got)
	}
	// No settings file: nothing to configure, no error.
	dir = writeAndroidProject(t, "")
	if err := configureGradleCache(dir, GradleCache{URL: "http://c/"}); err != nil {
		t.Fatal(err)
	}
	if got := readFile(t, filepath.Join(dir, "android", "gradle.properties")); got != "android.useAndroidX=true" {
		t.Fatalf("gradle.properties changed without settings: %q", got)
	}
}

func TestFlutterExecutor_BuildEnvCarriesCacheToken(t *testing.T) {
	e := NewFlutterExecutor()
	if env := strings.Join(e.buildEnv(), "\n"); strings.Contains(env, gradleCacheTokenEnv+"=") {
		t.Fatal("token env set without a cache")
	}
	e.SetGradleCache(GradleCache{URL: "http://c/", Token: "tok"})
	if env := strings.Join(e.buildEnv(), "\n"); !strings.Contains(env, gradleCacheTokenEnv+"=tok") {
		t.Fatal("token env missing")
	}
}
