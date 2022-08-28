// Copyright 2022 The Venice Authors. All rights reserved.
// Use of this source code is governed by an MIT-style license that can be
// found in the LICENSE file.

#include <stdarg.h>

#include "venice.h"

#define VENICE_LIST_MIN_CAPACITY 8

venice_list_t* venice_list_new(uint64_t capacity) {
  venice_list_t* list = venice_malloc(sizeof *list);
  list->length = 0;
  list->capacity =
      capacity < VENICE_LIST_MIN_CAPACITY ? VENICE_LIST_MIN_CAPACITY : capacity;
  list->items = venice_malloc((sizeof *list->items) * capacity);
  return list;
}

venice_list_t* venice_list_from_varargs(uint64_t length, ...) {
  venice_list_t* list = venice_malloc(sizeof *list);
  list->length = 0;

  uint64_t capacity =
      length < VENICE_LIST_MIN_CAPACITY ? VENICE_LIST_MIN_CAPACITY : length;
  list->capacity = capacity;
  list->items = venice_malloc((sizeof *list->items) * capacity);

  va_list args;
  va_start(args, length);

  for (uint64_t i = 0; i < length; i++) {
    uint64_t arg = va_arg(args, uint64_t);
    list->items[i] = arg;
    list->length++;
  }

  va_end(args);
  return list;
}

uint64_t venice_list_index(venice_list_t* list, uint64_t n) {
  if (n >= list->length) {
    runtime_error("index out of bounds");
  }
  return list->items[n];
}

void venice_list_resize(venice_list_t* list) {
  // TODO: what happens on overflow?
  uint64_t new_capacity = list->capacity * 2;
  list->capacity = new_capacity;
  list->items =
      venice_realloc(list->items, (sizeof *list->items) * new_capacity);
}

void venice_list_append(venice_list_t* list, uint64_t x) {
  if (list->length == list->capacity) {
    venice_list_resize(list);
  }
  list->items[list->length++] = x;
}
