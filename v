#!/bin/bash
# Compile and execute a Venice program in one step.
#
# The first argument is the path to the Venice program. All other args are
# passed on to the program.

set -eu

input_path=$1
shift
output_path=${input_path%.vn}

# Remove the executable on exit.
cleanup() {
  arg=$?
  rm -f "$output_path"
  exit $?
}

trap cleanup EXIT

# Compile the program.
cargo run -- "$input_path"

# Run the executable.
./"$output_path" "$@"
