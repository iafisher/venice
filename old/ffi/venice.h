#ifndef VENICE_H
#define VENICE_H

#include <stdlib.h>

typedef enum { VENICE_TYPE_INTEGER, VENICE_TYPE_LIST, VENICE_TYPE_STRING } VeniceType;

struct VeniceObject;

typedef struct {
    int value;
} VeniceIntObject;

typedef struct {
    const char* value;
} VeniceStringObject;

typedef struct {
    size_t length;
    size_t capacity;
    struct VeniceObject** items;
} VeniceListObject;

typedef struct VeniceObject {
    VeniceType vtype;
    union {
        VeniceIntObject* int_object;
        VeniceStringObject* string_object;
        VeniceListObject* list_object;
    } object;
} VeniceObject;

VeniceObject* NewVeniceIntObject(int);
VeniceObject* NewVeniceStringObject(const char*);
VeniceObject* NewVeniceListObject();
void VeniceListObjectAppend(VeniceObject* list, VeniceObject* obj);

void FreeVeniceObject(VeniceObject*);

#endif
