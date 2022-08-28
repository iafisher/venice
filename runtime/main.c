// Copyright 2022 The Venice Authors. All rights reserved.
// Use of this source code is governed by an MIT-style license that can be
// found in the LICENSE file.
//
// This is the entrypoint for all Venice programs. It does some initialization
// before passing control to the user-defined `venice_main` function.

#include "venice.h"

// Defined by the Venice program itself.
int venice_main();

int main(int argc, char* argv[]) {
  venice_list_t* list = venice_list_new(argc);

  while (*argv) {
    venice_string_t* s = venice_string_new(*argv);
    venice_list_append(list, (uint64_t)s);
    argv++;
  }

  return venice_main(list);
}
