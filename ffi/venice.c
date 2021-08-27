#include <stdio.h>
#include "venice.h"

VeniceObject* NewVeniceIntObject(int value) {
    VeniceObject* o = malloc(sizeof *o);
    o->vtype = VENICE_TYPE_INTEGER;
    o->value.int_value = value;
    return o;
}

VeniceObject* NewVeniceStringObject(const char* value) {
    VeniceObject* o = malloc(sizeof *o);
    o->vtype = VENICE_TYPE_STRING;
    o->value.string_value = value;
    return o;
}
