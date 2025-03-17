import subprocess
import os

# Set the environment variable inside the script
os.environ['DOCKER_DEFAULT_PLATFORM'] = 'linux/amd64'

# AWS ECR configuration
AWS_REGION = ''
AWS_ACCOUNT_ID = ''
IMAGES_FILE_PATH = "images.txt"  # Path to the images file

def run_command(command):
    """Run a shell command and return the output."""
    result = subprocess.run(command, shell=True, capture_output=True, text=True)
    if result.returncode != 0:
        print(f"Command failed: {command}\nError: {result.stderr}")
        raise Exception(result.stderr)
    return result.stdout.strip()

def ecr_repository_exists(repo_name):
    """Check if an ECR repository exists."""
    try:
        run_command(f"aws ecr describe-repositories --repository-names {repo_name} --region {AWS_REGION}")
        print(f"Repository {repo_name} already exists.")
        return True
    except Exception:
        print(f"Repository {repo_name} does not exist. It will be created.")
        return False

def create_ecr_repository(repo_name):
    """Create an ECR repository if it doesn't exist."""
    run_command(f"aws ecr create-repository --repository-name {repo_name} --region {AWS_REGION}")
    print(f"Created repository {repo_name}.")

def login_to_ecr():
    """Authenticate Docker to the ECR registry."""
    login_command = f"aws ecr get-login-password --region {AWS_REGION} | docker login --username AWS --password-stdin {AWS_ACCOUNT_ID}.dkr.ecr.{AWS_REGION}.amazonaws.com"
    run_command(login_command)
    print("Logged into AWS ECR.")

def push_image_to_ecr(source_image, target_image):
    """Tag and push the image to ECR."""
    run_command(f"docker pull {source_image}")
    run_command(f"docker tag {source_image} {target_image}")
    run_command(f"docker push {target_image}")
    print(f"Pushed {target_image} to ECR.")

def process_images():
    """Read images from file, check/create repos, and push images to ECR."""
    login_to_ecr()

    with open(IMAGES_FILE_PATH, 'r') as file:
        for line in file:
            source_image = line.strip()
            if source_image:
                # Extract repo path and tag (e.g., 'alpine/curl' from 'docker.io/alpine/curl:3.14')
                repo_path = '/'.join(source_image.split('/')[1:]).split(':')[0]
                tag = source_image.split(':')[-1]

                # Define ECR repository name and full ECR image path
                repo_name = repo_path
                ecr_image = f"{AWS_ACCOUNT_ID}.dkr.ecr.{AWS_REGION}.amazonaws.com/{repo_name}:{tag}"

                # Check if the repository exists, if not, create it
                if not ecr_repository_exists(repo_name):
                    create_ecr_repository(repo_name)

                # Push the image to ECR
                push_image_to_ecr(source_image, ecr_image)

if __name__ == "__main__":
    process_images()
