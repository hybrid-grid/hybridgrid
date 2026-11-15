Implemented task C0 without changing the frozen paths, switching branches, or committing.

**Files changed:** [parser.go](<D:\Đồ án tốt nghiệp\hybridgrid\internal\compiler\parser.go>), [parser_test.go](<D:\Đồ án tốt nghiệp\hybridgrid\internal\compiler\parser_test.go>), [preprocess.go](<D:\Đồ án tốt nghiệp\hybridgrid\internal\compiler\preprocess.go>), [preprocess_test.go](<D:\Đồ án tốt nghiệp\hybridgrid\internal\compiler\preprocess_test.go>), [build.go](<D:\Đồ án tốt nghiệp\hybridgrid\internal\cli\build\build.go>), [build_test.go](<D:\Đồ án tốt nghiệp\hybridgrid\internal\cli\build\build_test.go>), [main.go](<D:\Đồ án tốt nghiệp\hybridgrid\cmd\hgbuild\main.go>), [main_test.go](<D:\Đồ án tốt nghiệp\hybridgrid\cmd\hgbuild\main_test.go>), [Dockerfile.base](<D:\Đồ án tốt nghiệp\hybridgrid\test\stress\Dockerfile.base>), [cmake-smoke.sh](<D:\Đồ án tốt nghiệp\hybridgrid\test\stress\cmake-smoke.sh>), and the new [cmake-smoke project](<D:\Đồ án tốt nghiệp\hybridgrid\test\stress\cmake-smoke>) (`CMakeLists.txt`, six `.cc` files, one `.c` file, and `smoke.h`).

The parser keeps separate operands paired with their options. Dependency options go into `DepFlags`; local `-E` preprocessing receives them, while worker commands omit them and local include paths. Without `-MF`, preprocessing derives `obj/name.d` from `obj/name.o` and sets the object as the dependency target. Cache hits also regenerate the depfile. Preprocessed C++ is sent to the worker with an explicit C++ preprocessed language flag. `hgbuild install-wrappers DIR` installs persistent `hgcc` and `hgcxx` scripts.

`git diff --check` passed. I did not run Go, Docker, or the smoke script. Claude should run:

```sh
go test -race ./internal/compiler/... ./internal/cli/build/... ./internal/worker/executor/... ./cmd/hgbuild/...
```