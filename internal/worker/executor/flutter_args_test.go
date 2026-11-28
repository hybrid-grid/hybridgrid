package executor

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	pb "github.com/h3nr1-d14z/hybridgrid/gen/go/hybridgrid/v1"
)

func TestBuildFlutterArgs(t *testing.T) {
	tests := []struct {
		name   string
		output flutterOutputType
		cfg    *pb.FlutterConfig
		want   string
	}{
		{"defaults", flutterOutputAPK, &pb.FlutterConfig{}, "build apk --release"},
		{"one abi, split", flutterOutputAPK, &pb.FlutterConfig{TargetPlatforms: []string{"android-arm64"}, SplitPerAbi: true},
			"build apk --release --target-platform android-arm64 --split-per-abi"},
		{"several abis", flutterOutputAPK, &pb.FlutterConfig{TargetPlatforms: []string{"android-arm", "android-x64"}},
			"build apk --release --target-platform android-arm,android-x64"},
		{"unknown platform dropped", flutterOutputAPK, &pb.FlutterConfig{TargetPlatforms: []string{"android-arm64", "ios; rm -rf /"}},
			"build apk --release --target-platform android-arm64"},
		{"bundle ignores split", flutterOutputAppBundle, &pb.FlutterConfig{SplitPerAbi: true}, "build appbundle --release"},
		{"flavor and defines", flutterOutputAPK, &pb.FlutterConfig{Flavor: "prod", DartDefines: map[string]string{"B": "2", "A": "1"}},
			"build apk --release --flavor prod --dart-define A=1 --dart-define B=2"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := strings.Join(buildFlutterArgs(tt.output, "--release", tt.cfg), " ")
			if got != tt.want {
				t.Fatalf("args = %q\nwant   %q", got, tt.want)
			}
		})
	}
}

func TestCollectFlutterArtifactsDropsDuplicateCopies(t *testing.T) {
	workDir := t.TempDir()
	write := func(rel, content string) {
		path := filepath.Join(workDir, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	// Gradle's copy and Flutter's copy of the same APK, plus a distinct split APK.
	write("build/app/outputs/apk/release/app-release.apk", "universal")
	write("build/app/outputs/flutter-apk/app-release.apk", "universal")
	write("build/app/outputs/flutter-apk/app-arm64-v8a-release.apk", "arm64")

	list, archive, err := collectFlutterArtifacts(workDir, flutterOutputAPK)
	if err != nil {
		t.Fatal(err)
	}
	var paths []string
	for _, a := range list {
		paths = append(paths, a.Path)
	}
	want := "build/app/outputs/flutter-apk/app-arm64-v8a-release.apk build/app/outputs/flutter-apk/app-release.apk"
	if strings.Join(paths, " ") != want {
		t.Fatalf("paths = %v\nwant %s", paths, want)
	}
	if len(archive) == 0 {
		t.Fatal("empty archive")
	}
}

func TestFlutterGradleEnv(t *testing.T) {
	gradleOpts := func(env []string) []string {
		var out []string
		for _, kv := range env {
			if strings.HasPrefix(kv, "GRADLE_OPTS=") {
				out = append(out, kv)
			}
		}
		return out
	}
	got := gradleOpts(flutterGradleEnv([]string{"PATH=/bin"}))
	if len(got) != 1 || got[0] != "GRADLE_OPTS="+flutterGradleOpts {
		t.Fatalf("fresh env GRADLE_OPTS = %v", got)
	}
	got = gradleOpts(flutterGradleEnv([]string{"GRADLE_OPTS=-Xmx2g", "PATH=/bin"}))
	if len(got) != 1 || got[0] != "GRADLE_OPTS=-Xmx2g "+flutterGradleOpts {
		t.Fatalf("operator GRADLE_OPTS not kept: %v", got)
	}
}
