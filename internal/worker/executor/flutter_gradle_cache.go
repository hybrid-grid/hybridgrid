package executor

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

// gradleCacheTokenEnv carries the cache password to Gradle, so the token
// never lands in a file of the build's work dir.
const gradleCacheTokenEnv = "HG_GRADLE_CACHE_TOKEN"

// gradleCacheMarker tags the block appended to settings.gradle(.kts).
const gradleCacheMarker = "// hybridgrid: shared Gradle remote build cache"

// GradleCache points Flutter builds at a Gradle HTTP build cache (served
// by the coordinator under /cache/gradle/). A zero value disables it.
type GradleCache struct {
	URL   string
	Token string
}

// Enabled reports whether a cache URL is configured.
func (c GradleCache) Enabled() bool { return c.URL != "" }

// NormalizeGradleCacheURL validates an http(s) cache URL and returns it
// with a trailing slash, as Gradle expects. The URL is written into a
// Gradle script, so characters that could break out of a string literal
// are rejected rather than escaped.
func NormalizeGradleCacheURL(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", nil
	}
	u, err := url.Parse(raw)
	if err != nil {
		return "", fmt.Errorf("invalid gradle cache url: %w", err)
	}
	if (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		return "", errors.New("invalid gradle cache url: want http(s)://host:port/cache/gradle/")
	}
	if u.User != nil {
		return "", errors.New("invalid gradle cache url: pass credentials with --token, not in the url")
	}
	if strings.ContainsAny(raw, "'\"\\$`\n\r") {
		return "", errors.New("invalid gradle cache url: contains quote, backslash or $")
	}
	if !strings.HasSuffix(raw, "/") {
		raw += "/"
	}
	return raw, nil
}

// configureGradleCache turns on Gradle's build cache in the work dir's
// copy of the Android project and points it at the shared remote cache.
// Only the temporary copy is changed — never the worker's ~/.gradle — so
// other builds on the machine are unaffected. A project without an
// android/ settings file is left alone.
func configureGradleCache(workDir string, cache GradleCache) error {
	if !cache.Enabled() {
		return nil
	}
	androidDir := filepath.Join(workDir, "android")
	settings, kotlin := "", false
	for _, name := range []string{"settings.gradle.kts", "settings.gradle"} {
		p := filepath.Join(androidDir, name)
		if _, err := os.Stat(p); err == nil {
			settings, kotlin = p, strings.HasSuffix(name, ".kts")
			break
		}
	}
	if settings == "" {
		return nil
	}
	existing, err := os.ReadFile(settings)
	if err != nil {
		return fmt.Errorf("failed to read gradle settings: %w", err)
	}
	if strings.Contains(string(existing), gradleCacheMarker) {
		return nil
	}

	block := groovyCacheBlock(cache.URL)
	if kotlin {
		block = kotlinCacheBlock(cache.URL)
	}
	if err := appendFile(settings, "\n"+block); err != nil {
		return fmt.Errorf("failed to configure gradle cache: %w", err)
	}
	if err := appendFile(filepath.Join(androidDir, "gradle.properties"), "\norg.gradle.caching=true\n"); err != nil {
		return fmt.Errorf("failed to enable gradle caching: %w", err)
	}
	return nil
}

func groovyCacheBlock(u string) string {
	return gradleCacheMarker + `
buildCache {
    remote(HttpBuildCache) {
        url = '` + u + `'
        allowInsecureProtocol = true
        push = true
        def hgToken = System.getenv('` + gradleCacheTokenEnv + `')
        if (hgToken) {
            credentials {
                username = 'hybridgrid'
                password = hgToken
            }
        }
    }
}
`
}

func kotlinCacheBlock(u string) string {
	return gradleCacheMarker + `
buildCache {
    remote<HttpBuildCache> {
        url = uri("` + u + `")
        isAllowInsecureProtocol = true
        isPush = true
        System.getenv("` + gradleCacheTokenEnv + `")?.takeIf { it.isNotEmpty() }?.let { hgToken ->
            credentials {
                username = "hybridgrid"
                password = hgToken
            }
        }
    }
}
`
}

func appendFile(path, text string) error {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	if _, err := f.WriteString(text); err != nil {
		f.Close()
		return err
	}
	return f.Close()
}
