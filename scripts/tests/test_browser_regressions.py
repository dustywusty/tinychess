import os
from pathlib import Path
import subprocess
import tempfile
import unittest

ROOT = Path(__file__).resolve().parents[2]


class BrowserRegressionsTest(unittest.TestCase):
    def run_suites(self, failure="", ready=True):
        with tempfile.TemporaryDirectory() as directory:
            task_dir = Path(directory)
            commands = task_dir / "bin"
            commands.mkdir()
            stubs = {
                "go": '#!/bin/sh\nprintf "#!/bin/sh\\nexec sleep 30\\n" > "$3"\nchmod +x "$3"\n',
                "curl": "#!/bin/sh\nexit " + ("0" if ready else "1") + "\n",
                "pnpm": '''#!/usr/bin/env python3
import os, pathlib, sys, time
folder = pathlib.Path(os.environ["RUNNER_TEMP"])
if "playwright" not in sys.argv:
    time.sleep(30)
else:
    suite = "web" if "@yourmove/web" in sys.argv else "mobile"
    (folder / (suite + "-started")).touch()
    deadline = time.monotonic() + 3
    while not all((folder / (name + "-started")).exists() for name in ["web", "mobile"]):
        if time.monotonic() > deadline:
            sys.exit(99)
        time.sleep(0.01)
    (folder / (suite + "-finished")).touch()
    sys.exit(1 if suite == os.environ["FAIL_SUITE"] else 0)
''',
            }
            if not ready:
                stubs["sleep"] = "#!/bin/sh\nexit 0\n"
            for name, content in stubs.items():
                path = commands / name
                path.write_text(content)
                path.chmod(0o755)
            result = subprocess.run(["bash", "scripts/browser-regressions.sh"], cwd=ROOT,
                                    env={**os.environ, "PATH": str(commands) + os.pathsep + os.environ["PATH"],
                                         "RUNNER_TEMP": directory, "FAIL_SUITE": failure},
                                    capture_output=True, text=True, timeout=10)
            return result, {path.name for path in task_dir.iterdir()}

    def test_runs_both_suites_concurrently(self):
        result, files = self.run_suites()
        self.assertEqual(result.returncode, 0, result.stderr)
        self.assertTrue({"web-finished", "mobile-finished"} <= files)

    def test_either_failure_fails_the_step_but_waits_for_both(self):
        for failure in ["web", "mobile"]:
            with self.subTest(failure=failure):
                result, files = self.run_suites(failure)
                self.assertNotEqual(result.returncode, 0)
                self.assertTrue({"web-finished", "mobile-finished"} <= files)

    def test_server_readiness_failure_stops_before_tests(self):
        result, files = self.run_suites(ready=False)
        self.assertNotEqual(result.returncode, 0)
        self.assertIn("did not become ready", result.stderr)
        self.assertNotIn("web-started", files)
        self.assertNotIn("mobile-started", files)
