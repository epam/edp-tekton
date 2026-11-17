"""Write images.txt: unique container images in the rendered pipelines-library chart.

Renders the chart with default values, so only resources enabled in
.Values.pipelines.deployableResources are included.

Requires: helm, PyYAML, `helm dependency build charts/pipelines-library`.
Run from any directory: python3 hack/images/update_image_list.py
"""

import re
import subprocess
import sys
from pathlib import Path

import yaml

HELM_RELEASE_NAME = "edp-tekton"
SCRIPT_DIR = Path(__file__).resolve().parent
HELM_CHART_PATH = SCRIPT_DIR.parent.parent / "charts" / "pipelines-library"
IMAGES_FILE_PATH = SCRIPT_DIR / "images.txt"

DEFAULT_REGISTRY = "docker.io"
PARAM_VALUE_KEYS = ("default", "value")

_COMPONENT = r"[a-z0-9]+(?:(?:[._]|__|-+)[a-z0-9]+)*"
_REGISTRY = r"(?:[a-zA-Z0-9-]+(?:\.[a-zA-Z0-9-]+)*(?::[0-9]+)?/)?"
_TAG = r"(?::[A-Za-z0-9_][A-Za-z0-9_.-]{0,127})"
_DIGEST = r"(?:@sha256:[a-f0-9]{64})"
# Tag or digest is mandatory.
IMAGE_REF = re.compile(
    rf"{_REGISTRY}{_COMPONENT}(?:/{_COMPONENT})*(?:{_TAG}{_DIGEST}?|{_DIGEST})"
)


def run_helm_template():
    command = ["helm", "template", HELM_RELEASE_NAME, str(HELM_CHART_PATH)]
    result = subprocess.run(command, capture_output=True, text=True)
    if result.returncode != 0:
        sys.exit(f"helm template failed:\n{result.stderr}")
    return result.stdout


def normalize(ref):
    """Prefix docker.io when the first path component is not a registry host."""
    first, sep, _ = ref.partition("/")
    if sep and ("." in first or ":" in first or first == "localhost"):
        return ref
    return f"{DEFAULT_REGISTRY}/{ref}"


def candidates(node):
    """Yield `image` values and `default`/`value` of params named *image*."""
    if isinstance(node, dict):
        if isinstance(node.get("image"), str):
            yield node["image"]
        if "image" in str(node.get("name", "")).lower():
            for key in PARAM_VALUE_KEYS:
                if isinstance(node.get(key), str):
                    yield node[key]
        for value in node.values():
            yield from candidates(value)
    elif isinstance(node, list):
        for item in node:
            yield from candidates(item)


def extract_images(manifests):
    images = set()
    for doc in yaml.safe_load_all(manifests):
        if doc:
            images.update(
                normalize(ref.strip())
                for ref in candidates(doc)
                if IMAGE_REF.fullmatch(ref.strip())
            )
    return sorted(images)


def save_images_to_file(images, file_path):
    file_path.write_text("".join(f"{image}\n" for image in images))
    print(f"Saved {len(images)} unique images to {file_path}")


def main():
    images = extract_images(run_helm_template())
    save_images_to_file(images, IMAGES_FILE_PATH)


if __name__ == "__main__":
    main()
