package zutil

import (
	"syscall"
	"unsafe"
)

type ProtectVar []byte

func NewProtectVar(size int) (ProtectVar, error) {
	return syscall.Mmap(
		0, 0, size,
		syscall.PROT_READ|syscall.PROT_WRITE,
		syscall.MAP_ANON|syscall.MAP_PRIVATE,
	)
}
func (p ProtectVar) Free() error {
	return syscall.Munmap(p)
}
func (p ProtectVar) Readonly() error {
	return syscall.Mprotect(p, syscall.PROT_READ)
}
func (p ProtectVar) ReadWrite() error {
	return syscall.Mprotect(p, syscall.PROT_READ|syscall.PROT_WRITE)
}
func (p ProtectVar) Pointer() unsafe.Pointer {
	return unsafe.Pointer(&p[0])
}
