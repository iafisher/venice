// Copyright 2022 The Venice Authors. All rights reserved.
// Use of this source code is governed by an MIT-style license that can be
// found in the LICENSE file.
//
// The test suite for the runtime library. Since much of the Venice language
// depends on the runtime library, it is important that it is well-tested.
//
// To define a new test case:
//
//   1. Add a `test_XYZ` function in the "Test suite" section. It should include
//      at least one `ASSERT` statement and end with `global_tests_passed++`.
//   2. Call the function in `main`.

#include <stdio.h>
#include <string.h>
#include <unistd.h>

#include "internal.h"
#include "venice.h"

uint32_t global_tests_passed = 0;
uint32_t global_tests_failed = 0;

// Assert a condition is true with an optional message defined by a printf-style
// format string and the message itself.
#define ASSERT_MSG(e, format, msg)                                             \
  do {                                                                         \
    if (!(e)) {                                                                \
      global_tests_failed++;                                                   \
      fprintf(stderr, "assertion failed at %s, line %d in %s" format ": %s\n", \
              __FILE__, __LINE__, __func__, msg, #e);                          \
      return;                                                                  \
    }                                                                          \
  } while (0)

// Assert a condition is true.
#define ASSERT(e) ASSERT_MSG(e, "%s", "")

#define ASSERT_STRING_EQ(actual, expected)                                     \
  ASSERT_MSG(strcmp((actual), (expected)) == 0, " (got = \"%s\")", actual)

// Assert a condition is true in a loop (the error message will show the loop
// index).
#define ASSERT_IN_LOOP(e, i) ASSERT_MSG(e, " (loop index=%zu)", i)


/**
 * Test suite
 */

void test_list_from_varargs() {
  venice_list_t* list = venice_list_from_varargs(3, 10, 20, 30);
  ASSERT(venice_list_length(list) == 3);
  ASSERT(venice_list_capacity(list) >= 3);
  ASSERT(venice_list_index(list, 0) == 10);
  ASSERT(venice_list_index(list, 1) == 20);
  ASSERT(venice_list_index(list, 2) == 30);
  global_tests_passed++;
}

void test_list_append() {
  venice_list_t* list = venice_list_new(1);

  for (uint64_t i = 1; i <= 100; i++) {
    venice_list_append(list, i);
  }

  ASSERT(venice_list_length(list) == 100);
  ASSERT(venice_list_capacity(list) >= 100);

  for (uint64_t i = 1; i <= 100; i++) {
    ASSERT_IN_LOOP(venice_list_index(list, i - 1) == i, i);
  }

  global_tests_passed++;
}

void test_file_read_all() {
  venice_string_t* path = venice_string_new("test_resources/alphabet.txt");
  FILE* f = venice_file_open(path);
  // Choose a small buffer size to force the buffer to be re-allocated multiple
  // times.
  venice_string_t* string = venice_file_read_all_with_buffer_size(f, 4);
  ASSERT_STRING_EQ(string->data, "abcdefghijklmnopqrstuvwxyz\n");
  venice_file_close(f);
  global_tests_passed++;
}

void test_string_concat() {
  venice_string_t* left = venice_string_new("Hello");
  venice_string_t* right = venice_string_new(", world!");
  venice_string_t* result = venice_string_concat(left, right);
  ASSERT_STRING_EQ(result->data, "Hello, world!");

  venice_string_t* another = venice_string_new(" Lorem ipsum");
  venice_string_t* result2 = venice_string_concat(result, another);
  ASSERT_STRING_EQ(result2->data, "Hello, world! Lorem ipsum");

  global_tests_passed++;
}


/**
 * Test runner
 */

int main(int argc, char* argv[]) {
  if (argc != 1) {
    fprintf(stderr, "error: unexpected argument: %s\n", argv[1]);
    return 1;
  }

  test_list_from_varargs();
  test_list_append();
  test_file_read_all();
  test_string_concat();

  if (global_tests_failed > 0) {
    printf("\n");
    printf("FAILURE: %d of %d tests failed.\n", global_tests_failed,
           global_tests_passed + global_tests_failed);
    return 1;
  } else {
    printf("All %d tests passed.\n", global_tests_passed);
    return 0;
  }
}
