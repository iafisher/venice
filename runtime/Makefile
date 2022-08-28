# Courtesy of https://tech.davis-hansson.com/p/make/
SHELL := bash
.ONESHELL:
.SHELLFLAGS := -eu -o pipefail -c
.DELETE_ON_ERROR:
MAKEFLAGS += --warn-undefined-variables
MAKEFLAGS += --no-builtin-rules

FLAGS = -Wall -Werror -Wextra -fPIC -shared

all: libvenice.so libvenice-debug.so main.o main-debug.o

libvenice.so: venice.c
	gcc $(FLAGS) -o $@ $^

libvenice-debug.so: venice.c
	gcc $(FLAGS) -g -o $@ $^

main.o: main.c
	gcc $(FLAGS) -o $@ $^

main-debug.o: main.c
	gcc $(FLAGS) -g -o $@ $^

clean:
	rm -f *.so

.PHONY: clean
