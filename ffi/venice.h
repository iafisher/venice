#ifndef VENICE_H
#define VENICE_H

#include <stdlib.h>

typedef enum { VENICE_TYPE_INTEGER, VENICE_TYPE_STRING } VeniceType;

typedef struct {
    VeniceType vtype;
    union {
        int int_value;
        const char* string_value;
    } value;
} VeniceObject;

VeniceObject* NewVeniceIntObject(int value);
VeniceObject* NewVeniceStringObject(const char* value);

#endif
