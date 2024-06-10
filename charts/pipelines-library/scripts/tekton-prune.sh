#!/usr/bin/env bash

set -o errtrace
trap 'echo "error occurred on line ${LINENO}"; exit 1' ERR

verify() {
  echo 'Verify that kubectl is installed'
  if ! command -v kubectl &> /dev/null; then
    echo "kubectl could not be found"
    exit 1
  fi
  echo 'Ok'
  echo 'Verify that namespace variables are defined'
  if test -z "${NAMESPACE}"; then
    echo 'NAMESPACE env variable not defined.'
    exit 1
  fi
  echo 'Ok'
}

get_pipelinerun_to_file() {
  pods_file_path="$1"
  kubectl get -n "${NAMESPACE}" pipelinerun -o name > "${pods_file_path}"
}


get_active_pipelineruns() {
  kubectl get -n "${NAMESPACE}" pipelineruns \
    -o jsonpath='{.items[?(@.status.conditions[0].reason=="Running")].metadata.name}'
}


delete_lines_from_file() {
  file="$1"
  lines_to_delete="$2"
  for line in ${lines_to_delete[@]}; do
    sed -i "/${line}/d" "${file}"
  done
}

prune_resources() {
  resources_to_delete_file_path="$1"
  type="$2"
  resource_list=''
  while IFS= read -r line || [[ -n "${line}" ]]; do
    if ! test -z "${line}"; then resource_list="${resource_list} ${type}${line}"; fi
  done < "$resources_to_delete_file_path"
  if test -z "${resource_list// }"; then
    echo 'No resources to delete'
  else
    kubectl delete -n "${NAMESPACE}" ${resource_list} --force --grace-period=0;
  fi
}

main() {
  separator=';'
  pvc_owner_kind='PipelineRun'
  pipelinerun_to_delete_file_path='/tmp/runs-to-delete.txt'

  verify

  echo 'Get active pipelineruns'
  active_pipelineruns=$(get_active_pipelineruns)
  echo "Running pipelineruns: $active_pipelineruns"

  echo "Get pipelinerun list"
  get_pipelinerun_to_file "${pipelinerun_to_delete_file_path}"
  cat "${pipelinerun_to_delete_file_path}"

  echo 'Exclude running pipelineruns from deletion list':
  delete_lines_from_file "${pipelinerun_to_delete_file_path}" "${active_pipelineruns}"
  cat "${pipelinerun_to_delete_file_path}"

  echo 'Delete pipelineruns'
  prune_resources "${pipelinerun_to_delete_file_path}" ''
  echo 'Ok'

}

main
