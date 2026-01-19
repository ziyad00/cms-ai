import json
import unittest
from pathlib import Path

from cms_ai.spec_validator import validate_template_spec


class TestTemplateSpecValidatorUnit(unittest.TestCase):
    def test_valid_spec_has_no_errors(self) -> None:
        spec = json.loads(
            Path("tests/fixtures/valid_spec.json").read_text(encoding="utf-8")
        )
        errors = validate_template_spec(spec)
        self.assertEqual(errors, [])

    def test_overlap_is_reported(self) -> None:
        spec = json.loads(
            Path("tests/fixtures/invalid_spec_overlap.json").read_text(encoding="utf-8")
        )
        errors = validate_template_spec(spec)
        self.assertTrue(any("overlap" in e.message for e in errors), errors)


if __name__ == "__main__":
    unittest.main()
