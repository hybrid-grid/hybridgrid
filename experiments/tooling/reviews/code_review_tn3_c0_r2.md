Round-1 findings:

1. Preprocessing-option order — **RESOLVED**
2. Make-quoted synthesized targets — **RESOLVED**
3. Indirect dependency generation — **RESOLVED**
4. Clang option matching — **OPEN**: `-isystem-after<dir>` remains unrecognized because only the separate form is declared; Clang defines it as joined-or-separate. `-iwithsysroot` is also absent. [LLVM option definitions](https://llvm.googlesource.com/llvm-project/%2B/5004c59803fdeb3389d30439a6cc8b1ff874df0c/clang/include/clang/Driver/Options.td)
5. Listed side-output options — **RESOLVED**
6. Reused smoke directory — **RESOLVED**

New bugs:

- **MAJOR** — [parser.go:287](<D:/Đồ án tốt nghiệp/hybridgrid/internal/compiler/parser.go:287>): `-Wp,-MMD,file.d,-MP` joins every component after `-MMD` into the filename (`file.d,-MP`). GCC specifies that commas delimit distinct preprocessor arguments.
- **MAJOR** — [parser.go:219](<D:/Đồ án tốt nghiệp/hybridgrid/internal/compiler/parser.go:219>): side-output detection misses `-gsplit-dwarf` (`.dwo`) and `-ftest-coverage`/`--coverage` (`.gcno`), so those commands remain distributable while required artifacts are lost.

NO-GO