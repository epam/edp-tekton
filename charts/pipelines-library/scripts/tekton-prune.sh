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
  echo 'Verify that required environment variables are defined'
  if test -z "${NAMESPACE}"; then
    echo 'NAMESPACE env variable not defined.'
    exit 1
  fi
  if test -z "${RECENT_MINUTES}"; then
    echo 'RECENT_MINUTES env variable not defined.'
    exit 1
  fi
  echo 'Ok'
}

get_pipelinerun_pods_to_file() {
  pods_file_path="$1"
  kubectl get -n "${NAMESPACE}" pod -l tekton.dev/memberOf -o name > "${pods_file_path}"
}

get_pipelinerun_pvcs_to_file() {
  pvcs_file_path="$1"
  separator="$2"
  owner="$3"
  kubectl get -n "${NAMESPACE}" -o json $(kubectl get -n "${NAMESPACE}" pvc -o name) \
    | jq -r --arg separator "${separator}" --arg owner "${owner}" '.items[]
    | select(.metadata.ownerReferences[0].kind? == "\($owner)")
    | "\(.metadata.name)\($separator)\(.metadata.ownerReferences[0].name)"' > "${pvcs_file_path}"
}

get_active_pipelineruns() {
  kubectl get -n "${NAMESPACE}" pipelineruns \
    -o jsonpath='{.items[?(@.status.conditions[0].reason=="Running")].metadata.name}'
}

get_recent_pipelineruns() {
  minutes="$1"
  date_minus_minutes_iso8601=$(TZ=UTC date -u +"%FT%TZ" --date "-${minutes} min")
  kubectl get pipelineruns -n "${NAMESPACE}" -o json \
    | jq -r --arg d "${date_minus_minutes_iso8601}" \
      '.items[] | select (.status.completionTime? > $d ) | .metadata.name'
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
  pods_to_delete_file_path='/tmp/pods-to-delete.txt'
  pvcs_to_delete_file_path='/tmp/PVCs-to-delete.txt'

  verify

  echo 'Get active pipelineruns'
  active_pipelineruns=$(get_active_pipelineruns)
  echo "active pipelineruns: $active_pipelineruns"

  echo "Get pipelineruns completed recently (in the last ${RECENT_MINUTES} minutes)"
  recent_pipelineruns=$(get_recent_pipelineruns "${RECENT_MINUTES}")
  echo "recent pipelineruns: $recent_pipelineruns"

  echo 'Get pods that need to be deleted, pods with tekton.dev/memberOf label:'
  get_pipelinerun_pods_to_file "${pods_to_delete_file_path}"
  cat "${pods_to_delete_file_path}"

  echo 'Exclude pods of the active and recent pipelineruns from deletion list':
  delete_lines_from_file "${pods_to_delete_file_path}" "${active_pipelineruns}"
  delete_lines_from_file "${pods_to_delete_file_path}" "${recent_pipelineruns}"
  cat "${pods_to_delete_file_path}"

  echo 'Get PVCs that were used by pipelineruns and now need to be deleted:'
  get_pipelinerun_pvcs_to_file "${pvcs_to_delete_file_path}" "${separator}" "${pvc_owner_kind}"
  cat "${pvcs_to_delete_file_path}"

  echo 'Exclude PVCs of the active and recent pipelineruns from deletion list:'
  delete_lines_from_file "${pvcs_to_delete_file_path}" "${active_pipelineruns}"
  delete_lines_from_file "${pvcs_to_delete_file_path}" "${recent_pipelineruns}"
  cat "${pvcs_to_delete_file_path}"

  echo 'Remove owner info from PVCs list'
  sed -i "s,${separator}.*,," "${pvcs_to_delete_file_path}"

  echo 'Delete pods'
  prune_resources "${pods_to_delete_file_path}" ''
  echo 'Ok'

  echo 'Delete pvcs'
  prune_resources "${pvcs_to_delete_file_path}" 'pvc/'
  echo 'Ok'
}

main
