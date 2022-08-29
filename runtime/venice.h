// Copyright 2022 The Venice Authors. All rights reserved.
// Use of this source code is governed by an MIT-style license that can be
// found in the LICENSE file.
//
// This directory implements the Venice runtime library, a set of C functions
// that Venice programs use for low-level functionality that would be impossible
// to write in pure Venice.
//
// All objects other than integers and booleans are heap-allocated so that any
// Venice object can be represented as a 64-bit integer: either a primitive
// value or a pointer to a larger type. This makes the compiler and runtime
// library simpler at the cost of efficiency. Future versions of the compiler
// may allow different-sized types.
//
// Currently, the runtime library has no garbage collection and Venice programs
// leak any memory that they allocate.

#ifndef VENICE_H_
#define VENICE_H_

#include <stddef.h>
#include <stdint.h>
#include <stdio.h>

/***
 * Type definitions
 */

typedef int64_t venice_i64;

typedef struct {
  uint64_t length;
  uint64_t capacity;
  uint64_t* items;
} venice_list_t;

typedef struct {
  // `data` must be null-terminated for compatibility with C string functions.
  // `length` holds the number of bytes in the string, not including the null
  // terminator.
  uint64_t length;
  char* data;
} venice_string_t;


/***
 * String functions
 */

// Constructs a Venice string object from a raw string literal. A new buffer is
// allocated and the data is copied into it.
venice_string_t* venice_string_new(char* data);

// Returns the string's length in bytes. Does not include the null terminator.
uint64_t venice_string_length(venice_string_t* string);

// Returns the concatenation of `left` and `right`, which are left unmodified.
venice_string_t* venice_string_concat(venice_string_t* left,
                                      venice_string_t* right);


/***
 * List functions
 */

// Constructs an empty list with the given capacity.
venice_list_t* venice_list_new(uint64_t capacity);

// Constructs a Venice list object from a variadic list of arguments.
venice_list_t* venice_list_from_varargs(uint64_t length, ...);

// Returns the n'th element of the list.
uint64_t venice_list_index(venice_list_t* list, uint64_t n);

// Appends an object to the end of the list.
void venice_list_append(venice_list_t* list, uint64_t x);

// Retrieves length and capacity.
uint64_t venice_list_length(venice_list_t* list);
uint64_t venice_list_capacity(venice_list_t* list);


/***
 * Input and output
 */

// Prints a string followed by a newline to standard output.
void venice_println(venice_string_t* s);

// Prints a string to standard output. No trailing newline is printed.
void venice_print(venice_string_t* s);

// TODO(#146): support lines of arbitrary length
#define MAX_LINE_LENGTH 128

// Prints a prompt to standard output and reads a line from standard input.
venice_string_t* venice_input(venice_string_t* s);

// Prints an integer followed by a newline to standard output.
// TODO: remove printint once there's a better way to print integers.
void venice_printint(venice_i64 x);

// Opens a file and returns a handle that can be passed to other file functions.
FILE* venice_file_open(venice_string_t* path);

// Reads the entire contents of the file into a string and returns it.
venice_string_t* venice_file_read_all(FILE* f);

// Closes the file.
void venice_file_close(FILE* f);


/***
 * Miscellaneous program utilities
 */

void venice_panic(venice_string_t* message);

#endif // VENICE_H_
