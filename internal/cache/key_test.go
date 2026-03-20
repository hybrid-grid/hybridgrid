package cache

import (
	"testing"

	pb "github.com/h3nr1-d14z/hybridgrid/gen/go/hybridgrid/v1"
)

func TestFlutterCacheKey_Deterministic(t *testing.T) {
	config := &pb.FlutterConfig{
		BuildMode:   "release",
		Flavor:      "production",
		DartDefines: map[string]string{"ENV": "prod"},
	}

	key1 := FlutterCacheKey(config, "abc123", "3.24.0")
	key2 := FlutterCacheKey(config, "abc123", "3.24.0")
	if key1 != key2 {
		t.Errorf("same inputs produced different keys: %q vs %q", key1, key2)
	}
}

func TestFlutterCacheKey_DifferentPubspecHash(t *testing.T) {
	config := &pb.FlutterConfig{
		BuildMode:   "release",
		Flavor:      "production",
		DartDefines: map[string]string{},
	}

	key1 := FlutterCacheKey(config, "abc123", "3.24.0")
	key2 := FlutterCacheKey(config, "xyz789", "3.24.0")
	if key1 == key2 {
		t.Errorf("different pubspec hash produced same key: %q", key1)
	}
}

func TestFlutterCacheKey_DifferentBuildMode(t *testing.T) {
	config1 := &pb.FlutterConfig{
		BuildMode:   "debug",
		Flavor:      "",
		DartDefines: map[string]string{},
	}
	config2 := &pb.FlutterConfig{
		BuildMode:   "release",
		Flavor:      "",
		DartDefines: map[string]string{},
	}

	key1 := FlutterCacheKey(config1, "abc123", "3.24.0")
	key2 := FlutterCacheKey(config2, "abc123", "3.24.0")
	if key1 == key2 {
		t.Errorf("different build mode produced same key: %q", key1)
	}
}

func TestFlutterCacheKey_DifferentFlavor(t *testing.T) {
	config1 := &pb.FlutterConfig{
		BuildMode:   "release",
		Flavor:      "production",
		DartDefines: map[string]string{},
	}
	config2 := &pb.FlutterConfig{
		BuildMode:   "release",
		Flavor:      "staging",
		DartDefines: map[string]string{},
	}

	key1 := FlutterCacheKey(config1, "abc123", "3.24.0")
	key2 := FlutterCacheKey(config2, "abc123", "3.24.0")
	if key1 == key2 {
		t.Errorf("different flavor produced same key: %q", key1)
	}
}

func TestFlutterCacheKey_DifferentFlutterVersion(t *testing.T) {
	config := &pb.FlutterConfig{
		BuildMode:   "release",
		Flavor:      "",
		DartDefines: map[string]string{},
	}

	key1 := FlutterCacheKey(config, "abc123", "3.22.0")
	key2 := FlutterCacheKey(config, "abc123", "3.24.0")
	if key1 == key2 {
		t.Errorf("different Flutter version produced same key: %q", key1)
	}
}

func TestFlutterCacheKey_DifferentDartDefines(t *testing.T) {
	config1 := &pb.FlutterConfig{
		BuildMode:   "release",
		Flavor:      "",
		DartDefines: map[string]string{"API_URL": "https://prod.example.com"},
	}
	config2 := &pb.FlutterConfig{
		BuildMode:   "release",
		Flavor:      "",
		DartDefines: map[string]string{"API_URL": "https://staging.example.com"},
	}

	key1 := FlutterCacheKey(config1, "abc123", "3.24.0")
	key2 := FlutterCacheKey(config2, "abc123", "3.24.0")
	if key1 == key2 {
		t.Errorf("different Dart defines produced same key: %q", key1)
	}
}

func TestFlutterCacheKey_EmptyFlavorAndDefines(t *testing.T) {
	config := &pb.FlutterConfig{
		BuildMode:   "debug",
		Flavor:      "",
		DartDefines: map[string]string{},
	}

	key := FlutterCacheKey(config, "abc123", "3.24.0")
	if key == "" {
		t.Error("empty flavor and defines produced empty key")
	}
}

func TestFlutterCacheKey_NilConfig(t *testing.T) {
	key := FlutterCacheKey(nil, "abc123", "3.24.0")
	if key == "" {
		t.Error("nil config produced empty key")
	}

	keyNil := FlutterCacheKey(nil, "abc123", "3.24.0")
	keyEmpty := FlutterCacheKey(&pb.FlutterConfig{
		BuildMode:   "",
		Flavor:      "",
		DartDefines: map[string]string{},
	}, "abc123", "3.24.0")
	if keyNil != keyEmpty {
		t.Errorf("nil config and empty config produced different keys: %q vs %q", keyNil, keyEmpty)
	}
}

func TestFlutterCacheKey_DartDefinesDeterminism(t *testing.T) {
	config := &pb.FlutterConfig{
		BuildMode: "release",
		Flavor:    "",
		DartDefines: map[string]string{
			"BUILD_NUMBER": "42",
			"API_URL":      "https://example.com",
			"APP_NAME":     "myapp",
		},
	}

	var keys [5]string
	for i := 0; i < 5; i++ {
		keys[i] = FlutterCacheKey(config, "abc123", "3.24.0")
	}
	for i := 1; i < 5; i++ {
		if keys[i] != keys[0] {
			t.Errorf("iteration %d produced different key: got %q, want %q", i, keys[i], keys[0])
		}
	}
}
