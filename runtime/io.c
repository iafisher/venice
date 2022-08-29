// Copyright 2022 The Venice Authors. All rights reserved.
// Use of this source code is governed by an MIT-style license that can be
// found in the LICENSE file.

#include <stdio.h>
#include <string.h>

#include "internal.h"
#include "venice.h"

void venice_println(venice_string_t* s) {
  printf("%s\n", s->data);
}

void venice_print(venice_string_t* s) {
  printf("%s", s->data);
}

// TODO(#146): support lines of arbitrary length
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

FILE* venice_file_open(venice_string_t* path) {
  FILE* f = fopen(path->data, "r");
  if (f == NULL) {
    runtime_error("failed to open file");
  }
  return f;
}

#define FILE_READ_BUFFER_SIZE 4096

venice_string_t* venice_file_read_all(FILE* f) {
  return venice_file_read_all_with_buffer_size(f, FILE_READ_BUFFER_SIZE);
}

venice_string_t* venice_file_read_all_with_buffer_size(FILE* f,
                                                       size_t buffer_size) {
  uint64_t length = 0;
  char* buf = venice_malloc((sizeof *buf) * (buffer_size + 1));
  while (1) {
    size_t nread = fread(buf + length, sizeof *buf, buffer_size, f);
    length += nread;
    if (nread == buffer_size) {
      buf = venice_realloc(buf, (sizeof *buf) * (length + buffer_size + 1));
    } else {
      if (ferror(f)) {
        runtime_error("failed to read from file");
      } else {
        break;
      }
    }
  }

  buf[length] = '\0';
  return venice_string_new_no_alloc(length, buf);
}

void venice_file_close(FILE* f) {
  fclose(f);
}
