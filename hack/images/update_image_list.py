import subprocess
import re

HELM_RELEASE_NAME = "edp-tekton"  # Specify your Helm release name
HELM_CHART_PATH = "charts/pipelines-library"   # Specify the path to your Helm chart
IMAGES_FILE_PATH = "images.txt"          # Output file for unique images

def run_helm_template():
    """Run 'helm template' and return the output as a string."""
    command = f"helm template {HELM_RELEASE_NAME} {HELM_CHART_PATH}"
    result = subprocess.run(command, shell=True, capture_output=True, text=True)
    if result.returncode != 0:
        print(f"Error running helm template: {result.stderr}")
        raise Exception("Failed to run helm template")
    return result.stdout

def extract_docker_images(manifests):
    """Extract all unique images from 'docker.io' registry in the manifests."""
    # Regex pattern to match images from docker.io
    docker_image_pattern = re.compile(r"docker\.io/[a-zA-Z0-9._/-]+:[a-zA-Z0-9._-]+")
    images = set(docker_image_pattern.findall(manifests))  # Use a set to remove duplicates
    return sorted(images)  # Sort the images alphabetically

def save_images_to_file(images, file_path):
    """Save images to a text file, one per line."""
    with open(file_path, 'w') as file:
        for image in images:
            file.write(f"{image}\n")
    print(f"Saved {len(images)} unique images to {file_path}")

def main():
    # Run helm template and get the manifests
    manifests = run_helm_template()

    # Extract images from docker.io registry
    images = extract_docker_images(manifests)

    # Save images to the images.txt file
    save_images_to_file(images, IMAGES_FILE_PATH)

if __name__ == "__main__":
    main()
