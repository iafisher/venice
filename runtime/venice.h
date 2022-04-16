#ifndef VENICE_H
#define VENICE_H

#include <stdbool.h>
#include <stddef.h>
#include <stdint.h>

// Type aliases
typedef bool venice_bool_t;
typedef int32_t venice_int_t;

// String type
typedef struct {
    char* data;
    size_t length;
} venice_string_t;

venice_string_t venice_string_new(const char*);

// Global functions
void venice_print(venice_string_t);

#endif // VENICE_H
