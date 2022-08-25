#!/bin/bash

set -eu

input_path=$1
output_path=${input_path%.vn}

cleanup() {
  arg=$?
  rm -f "$output_path"
  exit $?
}

trap cleanup EXIT

cargo run -- "$input_path"
./"$output_path"
