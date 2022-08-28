// Copyright 2022 The Venice Authors. All rights reserved.
// Use of this source code is governed by an MIT-style license that can be
// found in the LICENSE file.

#include <string.h>

#include "venice.h"

venice_string_t* venice_string_new(char* data) {
  venice_string_t* s = venice_malloc(sizeof *s);
  s->length = strlen(data);
  s->data = venice_malloc((sizeof *s->data) * s->length);
  memcpy(s->data, data, s->length + 1);
  return s;
}
