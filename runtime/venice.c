// Copyright 2022 The Venice Authors. All rights reserved.
// Use of this source code is governed by an MIT-style license that can be
// found in the LICENSE file.
//
// This file implements the Venice runtime library, a set of C functions that Venice
// programs use for low-level functionality that would be impossible to write in pure
// Venice.

#include <stdarg.h>
#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

typedef int64_t venice_i64;

typedef struct {
  size_t length;
  uint64_t* items;
} venice_list_t;

typedef struct {
  // `data` must be null-terminated for compatibility with C string functions. `length`
  // holds the number of bytes in the string, not including the null terminator.
  size_t length;
  char* data;
} venice_string_t;

static void* venice_malloc(size_t size) {
  void* ptr = malloc(size);
  if (ptr == NULL) {
    fputs("runtime error: out of memory\n", stderr);
    exit(EXIT_FAILURE);
  }
  return ptr;
}

extern void venice_println(venice_string_t* s) {
  printf("%s\n", s->data);
}

// TODO: remove printint once there's a better way to print integers.
extern void venice_printint(venice_i64 x) {
  printf("%ld\n", x);
}

extern venice_string_t* venice_string_new(char* data) {
  venice_string_t* s = venice_malloc(sizeof *s);
  s->length = strlen(data);
  s->data = venice_malloc((sizeof *s->data) * s->length);
  memcpy(s->data, data, s->length + 1);
  return s;
}

extern venice_list_t* venice_list_new(size_t length, ...) {
  venice_list_t* list = venice_malloc(sizeof *list);
  list->length = length;
  list->items = venice_malloc((sizeof *list->items) * length);

  va_list args;
  va_start(args, length);

  for (size_t i = 0; i < length; i++) {
    uint64_t arg = va_arg(args, uint64_t);
    list->items[i] = arg;
  }

  va_end(args);
  return list;
}

extern uint64_t venice_list_index(venice_list_t* list, size_t index) {
  return list->items[index];
}
