# Courtesy of https://tech.davis-hansson.com/p/make/
SHELL := bash
.ONESHELL:
.SHELLFLAGS := -eu -o pipefail -c
.DELETE_ON_ERROR:
MAKEFLAGS += --warn-undefined-variables
MAKEFLAGS += --no-builtin-rules

FLAGS = -Wall -Werror -Wextra -fPIC -shared
SRCS = internal.c io.c list.c string.c

all: libvenice.so libvenice-debug.so main.o main-debug.o

libvenice.so: $(SRCS) venice.h
	gcc $(FLAGS) -o $@ $^

libvenice-debug.so: $(SRCS) venice.h
	gcc $(FLAGS) -g -o $@ $^

main.o: main.c venice.h
	gcc $(FLAGS) -o $@ $^

main-debug.o: main.c venice.h
	gcc $(FLAGS) -g -o $@ $^

clean:
	rm -f *.so *.o

.PHONY: clean
