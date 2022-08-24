#include <unistd.h>

extern size_t venice_string_length(const char* s) {
  const char* p = s;
  while (*s) {
    s++;
  }
  return s - p;
}

extern void venice_println(const char* s) {
  write(STDOUT_FILENO, s, venice_string_length(s));

  const char* newline = "\n";
  write(STDOUT_FILENO, newline, 1);
}
