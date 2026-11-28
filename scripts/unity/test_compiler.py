import json
import os
from pathlib import Path
import shutil
import subprocess
import tempfile
import unittest


class CompilerRoutingTest(unittest.TestCase):
    def test_compiler_routes(self):
        with tempfile.TemporaryDirectory() as tmp:
            root = Path(tmp)
            frontend = root / "clang++"
            shutil.copy(Path(__file__).with_name("hg-unity-il2cpp"), frontend)
            originals = root / "original"
            originals.mkdir()
            backend = root / "hgbuild"
            for path in [backend, originals / "clang++"]:
                path.write_text("#!/usr/bin/env python3\nimport json,sys\nprint(json.dumps(sys.argv))\n")
                path.chmod(0o755)
            env = dict(os.environ, HG_UNITY_ORIGINAL_BIN=str(originals),
                       HG_UNITY_HGBUILD=str(backend), HG_UNITY_BUILD_MARKER=str(root / "compiled"))
            flags = ["-c", "il2cppOutput/cpp/Assembly CSharp.cpp", "-o", "part.o"]
            (root / "args.rsp").write_text('-c "il2cppOutput/cpp/Assembly CSharp.cpp" -o part.o')
            def invoke(args):
                return json.loads(subprocess.check_output([frontend, *args], cwd=root, env=env))
            self.assertEqual(invoke(["@args.rsp"]), [str(backend), "c++", "--arch=x86_64", *flags])
            for args in [["part.o", "-o", "game"], ["-c", "plugin.cpp"], ["--version"]]:
                self.assertEqual(invoke(args), [str(originals / "clang++"), *args])


if __name__ == "__main__":
    unittest.main()
