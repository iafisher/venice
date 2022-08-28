// Copyright 2022 The Venice Authors. All rights reserved.
// Use of this source code is governed by an MIT-style license that can be
// found in the LICENSE file.

#include <stdio.h>
#include <stdlib.h>

#include "venice.h"

void runtime_error(const char* message) {
  fprintf(stderr, "runtime error: %s\n", message);
  exit(EXIT_FAILURE);
}

void* venice_malloc(size_t size) {
  void* ptr = malloc(size);
  if (ptr == NULL) {
    runtime_error("out of memory");
  }
  return ptr;
}

void* venice_realloc(void* ptr, size_t new_size) {
  void* ret = realloc(ptr, new_size);
  if (ret == NULL) {
    free(ptr);
    runtime_error("out of memory");
  }
  return ptr;
}
