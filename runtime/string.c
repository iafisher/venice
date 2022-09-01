// Copyright 2022 The Venice Authors. All rights reserved.
// Use of this source code is governed by an MIT-style license that can be
// found in the LICENSE file.

#include <string.h>

#include "internal.h"
#include "venice.h"

venice_string_t* venice_string_new(char* data) {
  venice_string_t* s = venice_malloc(sizeof *s);
  s->length = strlen(data);
  s->data = venice_malloc((sizeof *s->data) * (s->length + 1));
  memcpy(s->data, data, s->length + 1);
  return s;
}

venice_string_t* venice_string_new_no_alloc(uint64_t length, char* data) {
  venice_string_t* s = venice_malloc(sizeof *s);
  s->length = length;
  s->data = data;
  return s;
}

venice_string_t* venice_string_concat(venice_string_t* left,
                                      venice_string_t* right) {
  uint64_t length = left->length + right->length + 1;
  char* data = venice_malloc((sizeof *data) * length);
  memcpy(data, left->data, left->length);
  memcpy(data + left->length, right->data, right->length + 1);
  return venice_string_new_no_alloc(length - 1, data);
}

uint64_t venice_string_length(venice_string_t* string) {
  return string->length;
}
