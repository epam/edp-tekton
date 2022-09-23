import tempfile
import yaml
import os
import json
from subprocess import check_output


def kind_builder(helm_chart):
    results = {}
    for resource in helm_chart:
        if resource:
            kind = resource["kind"].lower()
            if kind not in results:
                results[kind] = {}
            results[kind][resource["metadata"]["name"]] = resource
    return results


def helm_template(config):
    with tempfile.NamedTemporaryFile() as temp:
        with open(temp.name, "w") as values:
            values.write(config)
        helm_cmd = f"helm template release-name -f {temp.name} ./charts/pipelines-library --namespace=ns"
        helm_chart = yaml.load_all(check_output(helm_cmd.split()), Loader=yaml.FullLoader)

        return kind_builder(helm_chart)
