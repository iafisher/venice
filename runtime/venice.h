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
 * Internal utility functions
 */

// Prints an error message and terminates the program.
void runtime_error(const char* message);

// Wrappers that exit the program if memory allocation fails.
void* venice_malloc(size_t size);
void* venice_realloc(void* ptr, size_t new_size);


/***
 * String functions
 */

// Constructs a Venice string object from a raw string literal.
venice_string_t* venice_string_new(char* data);


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


/***
 * Input and output
 */

// Prints a string followed by a newline to standard output.
void venice_println(venice_string_t* s);

// Prints a string to standard output. No trailing newline is printed.
void venice_print(venice_string_t* s);

// TODO: support lines of arbitrary length
#define MAX_LINE_LENGTH 128

// Prints a prompt to standard output and reads a line from standard input.
venice_string_t* venice_input(venice_string_t* s);

// Prints an integer followed by a newline to standard output.
// TODO: remove printint once there's a better way to print integers.
void venice_printint(venice_i64 x);

#endif // VENICE_H_
