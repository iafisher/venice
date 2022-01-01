#include "venice.h"

extern VeniceObject* return42(VeniceObject* args) {
    return NewVeniceIntObject(42);
}

extern VeniceObject* double_it(VeniceObject* args) {
    VeniceIntObject* int_obj = args->object.list_object->items[0]->object.int_object;
    return NewVeniceIntObject(int_obj->value * 2);
}

extern VeniceObject* return42string(VeniceObject* args) {
    return NewVeniceStringObject("42");
}
