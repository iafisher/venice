// Copyright 2022 The Venice Authors. All rights reserved.
// Use of this source code is governed by an MIT-style license that can be
// found in the LICENSE file.

#include <stdio.h>
#include <stdlib.h>

#include "venice.h"

void venice_panic(venice_string_t* message) {
  fprintf(stderr, "panic: %s\n", message->data);
  exit(EXIT_FAILURE);
}
