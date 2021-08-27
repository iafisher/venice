#include "venice.h"

extern VeniceObject* return42(VeniceObject* args) {
    return NewVeniceIntObject(42);
}

extern VeniceObject* double_it(VeniceObject* args) {
    return NewVeniceIntObject(args->value.int_value * 2);
}

extern VeniceObject* return42string(VeniceObject* args) {
    return NewVeniceStringObject("42");
}
