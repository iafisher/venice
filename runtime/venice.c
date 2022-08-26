#include <stdint.h>
#include <stdio.h>

typedef int64_t venice_i64;

extern void venice_println(const char* s) {
  printf("%s\n", s);
}

// TODO: remove printint once there's a better way to print integers.
extern void venice_printint(venice_i64 x) {
  printf("%ld\n", x);
}
