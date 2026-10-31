import os
from pathlib import Path
import subprocess
import tempfile
import unittest


SCRIPT = Path(__file__).resolve().parents[2] / "deploy/ovh/ci-command.sh"
IMAGE = "ghcr.io/dustywusty/tinychess@sha256:" + "a" * 64
COMMIT = "b" * 40


class RestrictedDeploymentCommandTests(unittest.TestCase):
    def run_command(self, command):
        with tempfile.TemporaryDirectory() as directory:
            # A valid request must forward literal arguments without running sudo.
            sudo = Path(directory) / "sudo"
            sudo.write_text('#!/bin/sh\nprintf "%s\\n" "$@"\n')
            sudo.chmod(0o755)
            return subprocess.run(
                ["bash", str(SCRIPT)],
                env={**os.environ, "PATH": directory + os.pathsep + os.environ["PATH"],
                     "SSH_ORIGINAL_COMMAND": command},
                capture_output=True, text=True,
            )

    def test_pinned_images_are_forwarded_as_arguments(self):
        result = self.run_command(f"deploy {IMAGE} {IMAGE} {COMMIT}")
        self.assertEqual(result.returncode, 0, result.stderr)
        self.assertEqual(result.stdout.splitlines(),
                         ["-n", "/opt/yourmove/deploy.sh", IMAGE, IMAGE, COMMIT])

    def test_shell_commands_tags_and_other_repositories_are_rejected(self):
        commands = [
            "", "bash", "id", "deploy", f"deploy {IMAGE} {IMAGE} {COMMIT};id",
            f"deploy {IMAGE} {IMAGE} {COMMIT} && id",
            f"deploy ghcr.io/dustywusty/tinychess:edge {IMAGE} {COMMIT}",
            f"deploy {IMAGE.replace('dustywusty', 'other')} {IMAGE} {COMMIT}",
            f"deploy {IMAGE} {IMAGE} $(id)",
        ]
        for command in commands:
            with self.subTest(command=command):
                result = self.run_command(command)
                self.assertNotEqual(result.returncode, 0)
                self.assertEqual(result.stdout, "")


if __name__ == "__main__":
    unittest.main()
