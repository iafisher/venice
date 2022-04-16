#include <stdio.h>
#include <stdlib.h>
#include <string.h>

#include "venice.h"

// Forward declarations
void* allocate(size_t sz, size_t count);

venice_string_t venice_string_new(const char* s) {
    size_t n = strlen(s) + 1;
    char* copied = allocate(sizeof *copied, n);
    memcpy(copied, s, n);
    venice_string_t r = {
        .data = copied,
        .length = n - 1,
    };
    return r;
}

void venice_print(venice_string_t s) {
    printf("%s\n", s.data);
}

void* allocate(size_t sz, size_t count) {
    void* data = malloc(count * sz);
    if (data == NULL) {
        fputs("Error: out of memory", stderr);
        exit(101);
    }
    return data;
}
