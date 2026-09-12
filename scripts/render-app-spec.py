#!/usr/bin/env python3
"""Replace source builds with immutable images while preserving the app configuration."""
import argparse
import copy
import re
from pathlib import Path

import yaml


def image_source(reference):
    match = re.fullmatch(r"ghcr\.io/([a-z0-9][a-z0-9-]*)/([a-z0-9][a-z0-9._/-]*)@(sha256:[a-f0-9]{64})", reference)
    if not match:
        raise ValueError("Use a GHCR image reference with a sha256 digest, not a tag.")
    owner, repository, digest = match.groups()
    return {"registry_type": "GHCR", "registry": owner, "repository": repository, "digest": digest}


def render(spec, backend, frontend):
    result = copy.deepcopy(spec)
    components = {}
    for group, name in [("services", "tinychess"), ("static_sites", "frontend")]:
        matches = [component for component in result.get(group, []) if component.get("name") == name]
        if len(matches) != 1:
            raise ValueError(f"Expected exactly one {group} component named {name}.")
        components[name] = matches[0]
    backend_component = components["tinychess"]
    for key in ["github", "git", "gitlab", "bitbucket", "source_dir", "dockerfile_path", "build_command", "environment_slug"]:
        backend_component.pop(key, None)
    backend_component["image"] = image_source(backend)

    # AppStaticSiteSpec rejects image sources. Its Dockerfile only extracts /site.
    image_source(frontend)
    frontend_component = components["frontend"]
    frontend_component["dockerfile_path"] = "Dockerfile.frontend-prebuilt"
    frontend_component.pop("build_command", None)
    envs = [env for env in frontend_component.get("envs", []) if env["key"] != "FRONTEND_IMAGE"]
    frontend_component["envs"] = envs + [{"key": "FRONTEND_IMAGE", "scope": "BUILD_TIME", "value": frontend}]
    return result


if __name__ == "__main__":
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--spec", default=".do/app.yaml")
    parser.add_argument("--backend", required=True)
    parser.add_argument("--frontend", required=True)
    parser.add_argument("--output", required=True)
    args = parser.parse_args()
    spec = yaml.safe_load(Path(args.spec).read_text())
    Path(args.output).write_text(yaml.safe_dump(render(spec, args.backend, args.frontend), sort_keys=False))
