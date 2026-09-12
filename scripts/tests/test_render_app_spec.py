import copy
import importlib.util
import unittest
from pathlib import Path

import yaml

ROOT = Path(__file__).resolve().parents[2]
spec = importlib.util.spec_from_file_location("render_app_spec", ROOT / "scripts/render-app-spec.py")
renderer = importlib.util.module_from_spec(spec)
spec.loader.exec_module(renderer)
BACKEND = "ghcr.io/dustywusty/tinychess@sha256:" + "a" * 64
FRONTEND = "ghcr.io/dustywusty/tinychess@sha256:" + "b" * 64


class RenderAppSpecTest(unittest.TestCase):
    def setUp(self):
        self.source = yaml.safe_load((ROOT / ".do/app.yaml").read_text())

    def test_preserves_database_domains_routes_and_runtime_settings(self):
        original = copy.deepcopy(self.source)
        rendered = renderer.render(self.source, BACKEND, FRONTEND)
        self.assertEqual(self.source, original)
        for key in self.source.keys() - {"services", "static_sites"}:
            self.assertEqual(rendered[key], self.source[key])
        backend = rendered["services"][0]
        expected = {k: v for k, v in self.source["services"][0].items()
                    if k not in {"github", "source_dir", "dockerfile_path"}}
        self.assertEqual({k: v for k, v in backend.items() if k != "image"}, expected)
        self.assertEqual(backend["image"], {"registry_type": "GHCR", "registry": "dustywusty",
                                           "repository": "tinychess", "digest": "sha256:" + "a" * 64})
        frontend = rendered["static_sites"][0]
        expected = {k: v for k, v in self.source["static_sites"][0].items() if k != "dockerfile_path"}
        self.assertEqual({k: v for k, v in frontend.items() if k not in {"dockerfile_path", "envs"}}, expected)
        self.assertNotIn("image", frontend)
        self.assertEqual(frontend["dockerfile_path"], "Dockerfile.frontend-prebuilt")
        self.assertEqual(frontend["envs"], [{"key": "FRONTEND_IMAGE", "scope": "BUILD_TIME", "value": FRONTEND}])
        self.assertEqual(rendered["services"][0]["instance_count"], 1)
        self.assertEqual(rendered["static_sites"][0]["output_dir"], "/site")

    def test_rejects_mutable_tags_and_malformed_references(self):
        for reference in ["ghcr.io/dustywusty/tinychess:edge", "", BACKEND + "\n", BACKEND[:-1],
                          BACKEND.replace("ghcr.io", "docker.io"), BACKEND.replace("sha256:", "sha512:")]:
            with self.subTest(reference=reference), self.assertRaises(ValueError):
                renderer.render(self.source, reference, FRONTEND)
            with self.subTest(reference=reference), self.assertRaises(ValueError):
                renderer.render(self.source, BACKEND, reference)

    def test_rejects_missing_or_duplicate_components(self):
        for group in ["services", "static_sites"]:
            for components in [[], self.source[group] * 2]:
                source = copy.deepcopy(self.source)
                source[group] = components
                with self.subTest(group=group), self.assertRaises(ValueError):
                    renderer.render(source, BACKEND, FRONTEND)

    def test_replaces_frontend_image_without_losing_other_build_variables(self):
        original = {"key": "OTHER", "scope": "BUILD_TIME", "value": "keep"}
        self.source["static_sites"][0]["envs"] = [original, {"key": "FRONTEND_IMAGE", "value": "old"}]
        self.source["static_sites"][0]["build_command"] = "obsolete build command"
        rendered = renderer.render(self.source, BACKEND, FRONTEND)
        envs = rendered["static_sites"][0]["envs"]
        self.assertEqual(envs, [original, {"key": "FRONTEND_IMAGE", "scope": "BUILD_TIME", "value": FRONTEND}])
        self.assertNotIn("build_command", rendered["static_sites"][0])
