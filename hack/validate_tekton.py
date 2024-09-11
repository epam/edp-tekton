import yaml
import sys
import tempfile
import os
from subprocess import check_output, CalledProcessError

PIPELINE_EXPECTED_ORDER = ['description', 'workspaces', 'params', 'results', 'tasks', 'finally']
PIPELINE_MANDATORY_FIELDS = ['description', 'params', 'tasks']

TASK_EXPECTED_ORDER = ['description', 'workspaces', 'params', 'results', 'steps']
TASK_MANDATORY_FIELDS = ['description', 'params', 'steps']

def helm_template(config, output_file):
    with tempfile.NamedTemporaryFile() as temp:
        with open(temp.name, "w") as values:
            values.write(config)
        helm_cmd = f"helm template release-name -f {temp.name} ./charts/pipelines-library --namespace=ns > {output_file}"
        try:
            check_output(helm_cmd, shell=True)
        except CalledProcessError as e:
            print(f"Error: Helm template command failed. {e}")
            sys.exit(1)

def check_description(yaml_file):
    try:
        with open(yaml_file, 'r') as file:
            documents = yaml.safe_load_all(file)
            for data in documents:
                kind = data.get('kind', '')
                metadata = data.get('metadata', {})
                name = metadata.get('name', 'Unnamed')
                spec = data.get('spec', {})

                if kind == 'Pipeline':
                    mandatory_fields = PIPELINE_MANDATORY_FIELDS
                    expected_order = PIPELINE_EXPECTED_ORDER
                elif kind == 'Task':
                    mandatory_fields = TASK_MANDATORY_FIELDS
                    expected_order = TASK_EXPECTED_ORDER
                else:
                    continue

                # Check for mandatory fields
                for field in mandatory_fields:
                    if field not in spec:
                        print(f"Error: '{kind}' named '{name}' does not have a '{field}' defined.")
                        return False

                if not check_key_order(spec, expected_order):
                    print(f"Error: '{kind}' named '{name}' does not have keys in the correct order.")
                    return False

                print(f"'{kind}' named '{name}' ok")
    except FileNotFoundError:
        print(f"Error: The file '{yaml_file}' does not exist.")
        return False
    except Exception as e:
        print(f"Error: Could not open or read the file '{yaml_file}'. {e}")
        return False

    return True

def check_key_order(spec, expected_order):
    keys = list(spec.keys())
    expected_keys = [key for key in expected_order if key in keys]

    # Check if all required keys are present
    for key in expected_order:
        if key not in keys and key not in ['workspaces', 'finally', 'results']:
            return False

    # Check if the order of the present keys is correct
    return keys[:len(expected_keys)] == expected_keys

if __name__ == "__main__":
    if len(sys.argv) != 2:
        print("Usage: python validate_tekton.py <config_file>")
        sys.exit(1)

    config_file = sys.argv[1]
    try:
        with open(config_file, 'r') as file:
            config = file.read()
    except FileNotFoundError:
        print(f"Error: The file '{config_file}' does not exist.")
        sys.exit(1)
    except Exception as e:
        print(f"Error: Could not open or read the file '{config_file}'. {e}")
        sys.exit(1)

    with tempfile.TemporaryDirectory() as temp_dir:
        output_file = os.path.join(temp_dir, "output.yaml")
        helm_template(config, output_file)
        if not check_description(output_file):
            sys.exit(1)
