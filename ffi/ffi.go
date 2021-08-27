package main

/*
#cgo LDFLAGS: -ldl -L /tmp/lol -lvenice
#include <dlfcn.h>
#include "venice.h"

typedef VeniceObject* (*venice_function_type)(VeniceObject*);

VeniceObject* venice_bridge(void* f, VeniceObject* args) {
	return ((venice_function_type)f)(args);
}
*/
import "C"

import "fmt"
import "unsafe"

func bytesToInt(bytes [8]byte) C.int {
	sum := 0
	multiplier := 1
	for _, b := range bytes {
		sum += multiplier * int(b)
		multiplier *= 256
	}
	return C.int(sum)
}

func main() {
	libraryName := "libreturn42.so"
	functionName := "double_it"

	args := C.VeniceObject{
		vtype: C.VENICE_TYPE_STRING,
		value: [8]byte{42, 0, 0, 0, 0, 0, 0, 0},
	}
	handle := C.dlopen(C.CString(libraryName), C.RTLD_LAZY)
	return42_ptr := C.dlsym(handle, C.CString(functionName))
	result := (*C.VeniceObject)(unsafe.Pointer(C.venice_bridge(return42_ptr, &args)))
	if result.vtype == C.VENICE_TYPE_INTEGER {
		fmt.Println(result.value)
		fmt.Println(bytesToInt(result.value))
	} else if result.vtype == C.VENICE_TYPE_STRING {
		pointerAsInt := bytesToInt(result.value)
		pointer := unsafe.Pointer(uintptr(pointerAsInt))
		// fmt.Println(C.CString(pointer))
		cstring := (*C.char)(pointer)
		fmt.Println(C.GoString(cstring))
	}
}
