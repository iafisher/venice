#include <stdio.h>
#include "venice.h"

VeniceObject* NewVeniceIntObject(int value) {
    VeniceIntObject* int_obj = malloc(sizeof *int_obj);
    int_obj->value = value;

    VeniceObject* obj = malloc(sizeof *obj);
    obj->vtype = VENICE_TYPE_INTEGER;
    obj->object.int_object = int_obj;
    return obj;
}

VeniceObject* NewVeniceStringObject(const char* value) {
    VeniceStringObject* str_obj = malloc(sizeof *str_obj);
    str_obj->value = value;

    VeniceObject* obj = malloc(sizeof *obj);
    obj->vtype = VENICE_TYPE_STRING;
    obj->object.string_object = str_obj;
    return obj;
}

VeniceObject* NewVeniceListObject() {
    VeniceListObject* list_obj = malloc(sizeof *list_obj);
    list_obj->length = 0;
    list_obj->capacity = 8;
    list_obj->items = malloc(list_obj->capacity * sizeof *list_obj->items);

    VeniceObject* obj = malloc(sizeof *obj);
    obj->vtype = VENICE_TYPE_LIST;
    obj->object.list_object = list_obj;
    return obj;
}

void VeniceListObjectAppend(VeniceObject* list_obj, VeniceObject* obj) {
    VeniceListObject* list = list_obj->object.list_object;
    if (list->length < list->capacity) {
        list->items[list->length++] = obj;
    }
}

void FreeVeniceObject(VeniceObject* obj) {
    switch (obj->vtype) {
        case VENICE_TYPE_INTEGER:
            free(obj->object.int_object);
            break;
        case VENICE_TYPE_LIST:
            for (size_t i = 0; i < obj->object.list_object->length; i++) {
                FreeVeniceObject(obj->object.list_object->items[i]);
            }

            free(obj->object.list_object->items);
            free(obj->object.list_object);
            break;
        case VENICE_TYPE_STRING:
            free(obj->object.string_object);
            break;
    }

    free(obj);
}
