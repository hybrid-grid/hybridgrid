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

func TestUnityCacheKey_Deterministic(t *testing.T) {
	config := &pb.UnityConfig{
		BuildMethod:      "BuildScript.Build",
		ScriptingBackend: "il2cpp",
		ExtraArgs:        map[string]string{"key": "value"},
	}

	key1 := UnityCacheKey(config, "abc123", "2022.3.10f1", pb.TargetPlatform_PLATFORM_ANDROID)
	key2 := UnityCacheKey(config, "abc123", "2022.3.10f1", pb.TargetPlatform_PLATFORM_ANDROID)
	if key1 != key2 {
		t.Errorf("same inputs produced different keys: %q vs %q", key1, key2)
	}
}

func TestUnityCacheKey_DifferentProjectHash(t *testing.T) {
	config := &pb.UnityConfig{
		BuildMethod:      "BuildScript.Build",
		ScriptingBackend: "il2cpp",
		ExtraArgs:        map[string]string{},
	}

	key1 := UnityCacheKey(config, "abc123", "2022.3.10f1", pb.TargetPlatform_PLATFORM_ANDROID)
	key2 := UnityCacheKey(config, "xyz789", "2022.3.10f1", pb.TargetPlatform_PLATFORM_ANDROID)
	if key1 == key2 {
		t.Errorf("different project hash produced same key: %q", key1)
	}
}

func TestUnityCacheKey_DifferentTargetPlatform(t *testing.T) {
	config := &pb.UnityConfig{
		BuildMethod:      "BuildScript.Build",
		ScriptingBackend: "il2cpp",
		ExtraArgs:        map[string]string{},
	}

	key1 := UnityCacheKey(config, "abc123", "2022.3.10f1", pb.TargetPlatform_PLATFORM_ANDROID)
	key2 := UnityCacheKey(config, "abc123", "2022.3.10f1", pb.TargetPlatform_PLATFORM_IOS)
	if key1 == key2 {
		t.Errorf("different target platform produced same key: %q", key1)
	}
}

func TestUnityCacheKey_DifferentScriptingBackend(t *testing.T) {
	config1 := &pb.UnityConfig{
		BuildMethod:      "BuildScript.Build",
		ScriptingBackend: "mono",
		ExtraArgs:        map[string]string{},
	}
	config2 := &pb.UnityConfig{
		BuildMethod:      "BuildScript.Build",
		ScriptingBackend: "il2cpp",
		ExtraArgs:        map[string]string{},
	}

	key1 := UnityCacheKey(config1, "abc123", "2022.3.10f1", pb.TargetPlatform_PLATFORM_ANDROID)
	key2 := UnityCacheKey(config2, "abc123", "2022.3.10f1", pb.TargetPlatform_PLATFORM_ANDROID)
	if key1 == key2 {
		t.Errorf("different scripting backend produced same key: %q", key1)
	}
}

func TestUnityCacheKey_DifferentBuildMethod(t *testing.T) {
	config1 := &pb.UnityConfig{
		BuildMethod:      "BuildScript.Build",
		ScriptingBackend: "il2cpp",
		ExtraArgs:        map[string]string{},
	}
	config2 := &pb.UnityConfig{
		BuildMethod:      "BuildScript.BuildWindows",
		ScriptingBackend: "il2cpp",
		ExtraArgs:        map[string]string{},
	}

	key1 := UnityCacheKey(config1, "abc123", "2022.3.10f1", pb.TargetPlatform_PLATFORM_ANDROID)
	key2 := UnityCacheKey(config2, "abc123", "2022.3.10f1", pb.TargetPlatform_PLATFORM_ANDROID)
	if key1 == key2 {
		t.Errorf("different build method produced same key: %q", key1)
	}
}

func TestUnityCacheKey_DifferentUnityVersion(t *testing.T) {
	config := &pb.UnityConfig{
		BuildMethod:      "BuildScript.Build",
		ScriptingBackend: "il2cpp",
		ExtraArgs:        map[string]string{},
	}

	key1 := UnityCacheKey(config, "abc123", "2022.3.10f1", pb.TargetPlatform_PLATFORM_ANDROID)
	key2 := UnityCacheKey(config, "abc123", "2023.1.0f1", pb.TargetPlatform_PLATFORM_ANDROID)
	if key1 == key2 {
		t.Errorf("different Unity version produced same key: %q", key1)
	}
}

func TestUnityCacheKey_DifferentExtraArgs(t *testing.T) {
	config1 := &pb.UnityConfig{
		BuildMethod:      "BuildScript.Build",
		ScriptingBackend: "il2cpp",
		ExtraArgs:        map[string]string{"API_URL": "https://prod.example.com"},
	}
	config2 := &pb.UnityConfig{
		BuildMethod:      "BuildScript.Build",
		ScriptingBackend: "il2cpp",
		ExtraArgs:        map[string]string{"API_URL": "https://staging.example.com"},
	}

	key1 := UnityCacheKey(config1, "abc123", "2022.3.10f1", pb.TargetPlatform_PLATFORM_ANDROID)
	key2 := UnityCacheKey(config2, "abc123", "2022.3.10f1", pb.TargetPlatform_PLATFORM_ANDROID)
	if key1 == key2 {
		t.Errorf("different extra args produced same key: %q", key1)
	}
}

func TestUnityCacheKey_ExtraArgsDeterminism(t *testing.T) {
	config := &pb.UnityConfig{
		BuildMethod:      "BuildScript.Build",
		ScriptingBackend: "il2cpp",
		ExtraArgs: map[string]string{
			"BUILD_NUMBER": "42",
			"API_URL":      "https://example.com",
			"APP_NAME":     "myapp",
		},
	}

	var keys [5]string
	for i := 0; i < 5; i++ {
		keys[i] = UnityCacheKey(config, "abc123", "2022.3.10f1", pb.TargetPlatform_PLATFORM_ANDROID)
	}
	for i := 1; i < 5; i++ {
		if keys[i] != keys[0] {
			t.Errorf("iteration %d produced different key: got %q, want %q", i, keys[i], keys[0])
		}
	}
}

func TestUnityCacheKey_NilConfig(t *testing.T) {
	key := UnityCacheKey(nil, "abc123", "2022.3.10f1", pb.TargetPlatform_PLATFORM_ANDROID)
	if key == "" {
		t.Error("nil config produced empty key")
	}

	keyNil := UnityCacheKey(nil, "abc123", "2022.3.10f1", pb.TargetPlatform_PLATFORM_ANDROID)
	keyEmpty := UnityCacheKey(&pb.UnityConfig{
		BuildMethod:      "",
		ScriptingBackend: "",
		ExtraArgs:        map[string]string{},
	}, "abc123", "2022.3.10f1", pb.TargetPlatform_PLATFORM_ANDROID)
	if keyNil != keyEmpty {
		t.Errorf("nil config and empty config produced different keys: %q vs %q", keyNil, keyEmpty)
	}
}
