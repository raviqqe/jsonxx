package codegen

import "llvm.org/llvm/bindings/go/llvm"

var genericPointerType = llvm.PointerType(llvm.Int8Type(), 0)
var thunkPointerType = genericPointerType

func toEntryName(s string) string {
	return s + "-entry"
}
