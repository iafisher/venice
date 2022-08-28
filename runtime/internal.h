// Copyright 2022 The Venice Authors. All rights reserved.
// Use of this source code is governed by an MIT-style license that can be
// found in the LICENSE file.
//
// This header declares internal functions which are used by the runtime but are
// not meant to be directly used by Venice programs.

#ifndef VENICE_INTERNAL_H_
#define VENICE_INTERNAL_H_

#include "venice.h"

// Prints an error message and terminates the program.
void runtime_error(const char* message);

// Wrappers that exit the program if memory allocation fails.
void* venice_malloc(size_t size);
void* venice_realloc(void* ptr, size_t new_size);

// Constructs a Venice string object without allocating a new buffer. The string
// takes ownership of `data`, which should not be used after calling this
// function. `data` should be null-terminated and `length` should be the length
// of the string not including the null terminator.
venice_string_t* venice_string_new_no_alloc(uint64_t length, char* data);

// Same as `venice_file_read_all` except that the buffer size is specified
// explicitly. This exists for convenience of testing and should not be used
// otherwise.
venice_string_t* venice_file_read_all_with_buffer_size(FILE* f,
                                                       size_t buffer_size);

#endif // VENICE_INTERNAL_H_
