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

import (
	"fmt"
	"unsafe"
)

func main() {
	libraryName := "libreturn42.so"
	functionName := "double_it"

	x := C.NewVeniceIntObject(21)
	args := C.NewVeniceListObject()
	C.VeniceListObjectAppend(args, x)

	handle := C.dlopen(C.CString(libraryName), C.RTLD_LAZY)
	fptr := C.dlsym(handle, C.CString(functionName))
	result := (*C.VeniceObject)(unsafe.Pointer(C.venice_bridge(fptr, args)))
	if result.vtype == C.VENICE_TYPE_INTEGER {
		intObject := *(**C.VeniceIntObject)(unsafe.Pointer(&result.object[0]))
		fmt.Println(intObject.value)
	} else if result.vtype == C.VENICE_TYPE_STRING {
		strObject := *(**C.VeniceStringObject)(unsafe.Pointer(&result.object[0]))
		strPtr := (*C.char)(strObject.value)
		fmt.Printf("%q\n", C.GoString(strPtr))
	}

	C.FreeVeniceObject(args)
	C.FreeVeniceObject(result)
}
