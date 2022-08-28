// Copyright 2022 The Venice Authors. All rights reserved.
// Use of this source code is governed by an MIT-style license that can be
// found in the LICENSE file.
//
// This file implements the Venice runtime library, a set of C functions that
// Venice programs use for low-level functionality that would be impossible to
// write in pure Venice.
//
// All objects other than integers and booleans are heap-allocated so that any
// Venice object can be represented as a 64-bit integer: either a primitive
// value or a pointer to a larger type. This makes the compiler and runtime
// library simpler at the cost of efficiency. Future versions of the compiler
// may allow different-sized types.
//
// Currently, the runtime library has no garbage collection and Venice programs
// leak any memory that they allocate.

#include <stdarg.h>
#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

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
static void runtime_error(const char* message) {
  fprintf(stderr, "runtime error: %s\n", message);
  exit(EXIT_FAILURE);
}

// A `malloc` wrapper that exits the program if `malloc` fails.
static void* venice_malloc(size_t size) {
  void* ptr = malloc(size);
  if (ptr == NULL) {
    runtime_error("out of memory");
  }
  return ptr;
}

static void* venice_realloc(void* ptr, size_t new_size) {
  void* ret = realloc(ptr, new_size);
  if (ret == NULL) {
    free(ptr);
    runtime_error("out of memory");
  }
  return ptr;
}

/***
 * String functions
 */

// Constructs a Venice string object from a raw string literal.
extern venice_string_t* venice_string_new(char* data) {
  venice_string_t* s = venice_malloc(sizeof *s);
  s->length = strlen(data);
  s->data = venice_malloc((sizeof *s->data) * s->length);
  memcpy(s->data, data, s->length + 1);
  return s;
}

/***
 * List functions
 */

#define VENICE_LIST_MIN_CAPACITY 8

// Constructs an empty list with the given capacity.
extern venice_list_t* venice_list_new(uint64_t capacity) {
  venice_list_t* list = venice_malloc(sizeof *list);
  list->length = 0;
  list->capacity =
      capacity < VENICE_LIST_MIN_CAPACITY ? VENICE_LIST_MIN_CAPACITY : capacity;
  list->items = venice_malloc((sizeof *list->items) * capacity);
  return list;
}

// Constructs a Venice list object from a variadic list of arguments.
extern venice_list_t* venice_list_from_varargs(uint64_t length, ...) {
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

// Returns the n'th element of the list.
extern uint64_t venice_list_index(venice_list_t* list, uint64_t n) {
  if (n >= list->length) {
    runtime_error("index out of bounds");
  }
  return list->items[n];
}

static void venice_list_resize(venice_list_t* list) {
  // TODO: what happens on overflow?
  uint64_t new_capacity = list->capacity * 2;
  list->capacity = new_capacity;
  list->items =
      venice_realloc(list->items, (sizeof *list->items) * new_capacity);
}

// Appends an object to the end of the list.
extern void venice_list_append(venice_list_t* list, uint64_t x) {
  if (list->length == list->capacity) {
    venice_list_resize(list);
  }
  list->items[list->length++] = x;
}

/***
 * Input and output
 */

// Prints a string followed by a newline to standard output.
extern void venice_println(venice_string_t* s) { printf("%s\n", s->data); }

// Prints a string to standard output. No trailing newline is printed.
extern void venice_print(venice_string_t* s) { printf("%s", s->data); }

// TODO: support lines of arbitrary length
#define MAX_LINE_LENGTH 128

// Prints a prompt to standard output and reads a line from standard input.
extern venice_string_t* venice_input(venice_string_t* s) {
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

// Prints an integer followed by a newline to standard output.
// TODO: remove printint once there's a better way to print integers.
extern void venice_printint(venice_i64 x) { printf("%ld\n", x); }
