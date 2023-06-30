package utils

import (
	"bytes"
	"reflect"
	"sync"
	"unsafe"
)

// var intSlicePool = sync.Pool{
// 	New: func() interface{} {
// 		ids := make([]int, 0, 16)
// 		return &ids
// 	},
// }

// var stringSlicePool = sync.Pool{
// 	New: func() interface{} {
// 		ids := make([]string, 0, 16)
// 		return &ids
// 	},
// }

// var interfacePool = sync.Pool{
// 	New: func() interface{} {
// 		ids := make([]interface{}, 0, 512)
// 		return &ids
// 	},
// }

// func GetIntSlice() []int {
// 	ids := intSlicePool.Get().(*[]int)
// 	*ids = (*ids)[:0]
// 	return *ids
// }

// func PutIntSlice(ids *[]int) {
// 	*ids = (*ids)[:0]
// 	intSlicePool.Put(ids)
// }

// func GetStringSlice() []string {
// 	ids := stringSlicePool.Get().(*[]string)
// 	*ids = (*ids)[:0]
// 	return *ids
// }

// func PutStringSlice(ids *[]string) {
// 	*ids = (*ids)[:0]
// 	stringSlicePool.Put(ids)
// }

// func GetInterfaceSlice() []interface{} {
// 	ids := interfacePool.Get().(*[]interface{})
// 	*ids = (*ids)[:0]
// 	return *ids
// }

// func PutInterfaceSlice(ids *[]interface{}) {
// 	*ids = (*ids)[:0]
// 	interfacePool.Put(ids)
// }

var bytesPool = sync.Pool{
	New: func() interface{} {
		var b []byte
		return bytes.NewBuffer(b)
	},
}

func GetBytes() *bytes.Buffer {
	return bytesPool.Get().(*bytes.Buffer)
}

func PutBytes(b *bytes.Buffer) {
	b.Reset()
	bytesPool.Put(b)
}

func String2bytes(s string) []byte {
	stringHeader := (*reflect.StringHeader)(unsafe.Pointer(&s))
	bh := reflect.SliceHeader{
		Data: stringHeader.Data,
		Len:  stringHeader.Len,
		Cap:  stringHeader.Len,
	}
	return *(*[]byte)(unsafe.Pointer(&bh))
}

func Bytes2string(b []byte) string {
	sliceHeader := (*reflect.SliceHeader)(unsafe.Pointer(&b))
	sh := reflect.StringHeader{
		Data: sliceHeader.Data,
		Len:  sliceHeader.Len,
	}
	return *(*string)(unsafe.Pointer(&sh))
}
