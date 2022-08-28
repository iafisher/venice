# Courtesy of https://tech.davis-hansson.com/p/make/
SHELL := bash
.ONESHELL:
.SHELLFLAGS := -eu -o pipefail -c
.DELETE_ON_ERROR:
MAKEFLAGS += --warn-undefined-variables
MAKEFLAGS += --no-builtin-rules

FLAGS = -Wall -Werror -Wextra
LIB_FLAGS = -fPIC -shared
SRCS = internal.c io.c list.c misc.c string.c
HEADERS = internal.h venice.h

all: libvenice.so libvenice-debug.so main.o main-debug.o

libvenice.so: $(SRCS) $(HEADERS)
	gcc $(FLAGS) $(LIB_FLAGS) -o $@ $^

libvenice-debug.so: $(SRCS) $(HEADERS)
	gcc $(FLAGS) $(LIB_FLAGS) -g -o $@ $^

main.o: main.c venice.h
	gcc $(FLAGS) -c -o $@ $<

main-debug.o: main.c venice.h
	gcc $(FLAGS) -g -c -o $@ $<

test: test.c libvenice-debug.so
# `-Wl,-R.` tells GCC to pass `-R.` to the linker, which causes the
# linker to look for dynamic libraries in the current directory at
# runtime.
#
# `-L. -l:libvenice-debug.so` tell the compiler/linker where to find
# the shared library at compile time.
	gcc $(FLAGS) -g -o $@ $< -Wl,-R. -L. -l:libvenice-debug.so

clean:
	rm -f *.so *.o test

.PHONY: clean
