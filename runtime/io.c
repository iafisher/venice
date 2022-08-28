// Copyright 2022 The Venice Authors. All rights reserved.
// Use of this source code is governed by an MIT-style license that can be
// found in the LICENSE file.

#include <stdio.h>
#include <string.h>

#include "venice.h"

void venice_println(venice_string_t* s) {
  printf("%s\n", s->data);
}

void venice_print(venice_string_t* s) {
  printf("%s", s->data);
}

// TODO: support lines of arbitrary length
#define MAX_LINE_LENGTH 128

venice_string_t* venice_input(venice_string_t* s) {
  printf("%s", s->data);
  fflush(stdout);

  char line[MAX_LINE_LENGTH];
  char* r = fgets(line, MAX_LINE_LENGTH, stdin);
  if (r == NULL) {
    runtime_error("fgets failed");
  }

  // Strip the trailing newline.
  size_t n = strlen(line);
  if (n > 0 && line[n - 1] == '\n') {
    line[n - 1] = '\0';
  }

  return venice_string_new(line);
}

void venice_printint(venice_i64 x) {
  printf("%ld\n", x);
}
