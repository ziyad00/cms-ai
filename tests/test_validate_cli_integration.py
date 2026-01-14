import subprocess
import sys
import unittest


class TestValidateCliIntegration(unittest.TestCase):
    def test_validate_cli_ok(self) -> None:
        result = subprocess.run(
            [sys.executable, "-m", "cms_ai.validate", "tests/fixtures/valid_spec.json"],
            capture_output=True,
            text=True,
        )
        self.assertEqual(result.returncode, 0, result.stderr)
        self.assertIn("OK", result.stdout)

    def test_validate_cli_fails_for_invalid(self) -> None:
        result = subprocess.run(
            [
                sys.executable,
                "-m",
                "cms_ai.validate",
                "tests/fixtures/invalid_spec_overlap.json",
            ],
            capture_output=True,
            text=True,
        )
        self.assertNotEqual(result.returncode, 0)
        self.assertIn("overlap", result.stderr)


if __name__ == "__main__":
    unittest.main()
