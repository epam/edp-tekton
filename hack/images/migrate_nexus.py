import subprocess
import os

# Set the environment variable inside the script
os.environ['DOCKER_DEFAULT_PLATFORM'] = 'linux/amd64'

# Nexus repository configuration
NEXUS_URL = ''  # Nexus URL (e.g., 'https://nexus.example.com')
NEXUS_REPOSITORY = 'mirror'  # Nexus repository name (e.g., 'docker-repository')
IMAGES_FILE_PATH = "images.txt"  # Path to the images file

def run_command(command):
    """Run a shell command and return the output."""
    result = subprocess.run(command, shell=True, capture_output=True, text=True)
    if result.returncode != 0:
        print(f"Command failed: {command}\nError: {result.stderr}")
        raise Exception(result.stderr)
    return result.stdout.strip()

def login_to_nexus():
    """Authenticate Docker to the Nexus registry."""
    login_command = f"docker login {NEXUS_URL} --username {os.getenv('NEXUS_USERNAME')} --password {os.getenv('NEXUS_PASSWORD')}"
    run_command(login_command)
    print("Logged into Nexus repository.")

def push_image_to_nexus(source_image, target_image):
    """Tag and push the image to Nexus."""
    run_command(f"docker pull {source_image}")
    run_command(f"docker tag {source_image} {target_image}")
    run_command(f"docker push {target_image}")
    print(f"Pushed {target_image} to Nexus.")

def process_images():
    """Read images from file, and push images to Nexus."""
    login_to_nexus()

    with open(IMAGES_FILE_PATH, 'r') as file:
        for line in file:
            source_image = line.strip()
            if source_image:
                # Extract repo path and tag (e.g., 'alpine/curl' from 'docker.io/alpine/curl:3.14')
                repo_path = '/'.join(source_image.split('/')[1:]).split(':')[0]
                tag = source_image.split(':')[-1]

                # Define Nexus repository URL and full Nexus image path
                nexus_image = f"{NEXUS_URL}/{NEXUS_REPOSITORY}/{repo_path}:{tag}"

                # Push the image to Nexus
                push_image_to_nexus(source_image, nexus_image)

if __name__ == "__main__":
    process_images()
