#!/bin/bash

export DISABLE_LOG=true
export IS_AGENT=true
/ko-app/k8s-offline "$@" 