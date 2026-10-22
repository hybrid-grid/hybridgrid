Fixed the six review findings on `feat/exp3` without committing.

- **Preprocessing order:** [parser.go](<D:/Đồ án tốt nghiệp/hybridgrid/internal/compiler/parser.go>) and [preprocess.go](<D:/Đồ án tốt nghiệp/hybridgrid/internal/compiler/preprocess.go>) now retain `-D`, `-U`, and include options in command-line order. The raw cache key also distinguishes different orders in [build.go](<D:/Đồ án tốt nghiệp/hybridgrid/internal/cli/build/build.go>).
- **Dependency targets:** synthesized targets use `-MQ`; explicit `-MT` and `-MQ` remain as supplied.
- **Indirect dependencies:** `-Wp,` and `-Xpreprocessor` dependency options now trigger depfile generation on cache misses and hits.
- **Option parsing:** an explicit form table handles the Clang include options and their operands.
- **Side outputs:** auxiliary, dump, and saved-temporary output requests stay local.
- **Smoke build:** [cmake-smoke.sh](<D:/Đồ án tốt nghiệp/hybridgrid/test/stress/cmake-smoke.sh>) uses a fresh temporary directory by default and removes a supplied build directory before configuring.

Regression coverage is in [parser_test.go](<D:/Đồ án tốt nghiệp/hybridgrid/internal/compiler/parser_test.go>), [preprocess_test.go](<D:/Đồ án tốt nghiệp/hybridgrid/internal/compiler/preprocess_test.go>), and [build_test.go](<D:/Đồ án tốt nghiệp/hybridgrid/internal/cli/build/build_test.go>), including explicit CPython-style local and remote argument lists. The existing `cpp-output` and link-test edits were preserved.

`git diff --check` passed. I did not run Go or Docker. Claude can run `go test -v -race ./internal/compiler/... ./internal/cli/build/...`; the attempted `sh -n` check could not start because Windows MSYS returned a `CreateFileMapping` error.