#!/usr/bin/env bash

# Copyright 2017 The Kubernetes Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

set -o errexit
set -o nounset
set -o pipefail

SCRIPT_DIR="$(dirname "${BASH_SOURCE[0]}")"
SCRIPT_ROOT="${SCRIPT_DIR}/.."
CODEGEN_PKG="${CODEGEN_PKG:-"${SCRIPT_ROOT}/"}"

source "${CODEGEN_PKG}/kube_codegen.sh"

THIS_PKG="gitee.com/we7coreteam/k8s-offline"

kube::codegen::gen_helpers \
    --boilerplate "${SCRIPT_ROOT}/hack/boilerplate.go.txt" \
    "${SCRIPT_ROOT}"

if [[ -n "${API_KNOWN_VIOLATIONS_DIR:-}" ]]; then
    report_filename="${API_KNOWN_VIOLATIONS_DIR}/codegen_violation_exceptions.list"
    if [[ "${UPDATE_API_KNOWN_VIOLATIONS:-}" == "true" ]]; then
        update_report="--update-report"
    fi
fi



# kube::codegen::gen_client \
#     --with-watch \
#     --with-applyconfig \
#     --output-dir "${SCRIPT_ROOT}/pkg/codegen/appcontroller" \
#     --output-pkg "${THIS_PKG}/pkg/codegen/appcontroller" \
#     --boilerplate "${SCRIPT_ROOT}/hack/boilerplate.go.txt" \
#     --one-input-api "appcontroller" \
#     "${SCRIPT_DIR}/../pkg/" 

# kube::codegen::gen_client \
#     --with-watch \
#     --with-applyconfig \
#     --output-dir "${SCRIPT_ROOT}/pkg/codegen/gpuclass" \
#     --output-pkg "${THIS_PKG}/pkg/codegen/gpuclass" \
#     --boilerplate "${SCRIPT_ROOT}/hack/boilerplate.go.txt" \
#     --one-input-api "gpuclass" \
#     "${SCRIPT_DIR}/../pkg/"  

# kube::codegen::gen_client \
#     --with-watch \
#     --with-applyconfig \
#     --output-dir "${SCRIPT_ROOT}/k8s/pkg/client/gpuclass" \
#     --output-pkg "${THIS_PKG}/k8s/pkg/client/gpuclass" \
#     --boilerplate "${SCRIPT_ROOT}/hack/boilerplate.go.txt" \
#     --one-input-api "gpuclass" \
#     "${SCRIPT_DIR}/../k8s/pkg/apis"  

# kube::codegen::gen_client \
#     --with-watch \
#     --with-applyconfig \
#     --output-dir "${SCRIPT_ROOT}/k8s/pkg/client/appgroup" \
#     --output-pkg "${THIS_PKG}/k8s/pkg/client/appgroup" \
#     --boilerplate "${SCRIPT_ROOT}/hack/boilerplate.go.txt" \
#     --one-input-api "appgroup" \
#     "${SCRIPT_DIR}/../k8s/pkg/apis"      

# kube::codegen::gen_client \
#     --with-watch \
#     --with-applyconfig \
#     --output-dir "${SCRIPT_ROOT}/k8s/pkg/client/mcpserver" \
#     --output-pkg "${THIS_PKG}/k8s/pkg/client/mcpserver" \
#     --boilerplate "${SCRIPT_ROOT}/hack/boilerplate.go.txt" \
#     --one-input-api "mcpserver" \
#     "${SCRIPT_DIR}/../k8s/pkg/apis"        

kube::codegen::gen_client \
    --with-watch \
    --with-applyconfig \
    --output-dir "${SCRIPT_ROOT}/k8s/pkg/client/microapp" \
    --output-pkg "${THIS_PKG}/k8s/pkg/client/microapp" \
    --boilerplate "${SCRIPT_ROOT}/hack/boilerplate.go.txt" \
    --one-input-api "microapp" \
    "${SCRIPT_DIR}/../k8s/pkg/apis"       