# 🐹 GoSneak: Advanced Evasion Framework in Go

**GoSneak** is a Go-native offensive security script for User-Mode Hook Evasion and Manual Memory Management. Implements Indirect Syscalls via custom ASM stubs to bypass EDR telemetry and utilizes RWX-transitioning memory pages outside of the Go Runtime's Garbage Collector (GC) heap. It bridges the gap between Go's high-level concurrency (Fibers/Goroutines) and low-level Win32/NTAPI interactions.


## 🔬 Technical Specifications

### 1. User-Mode Hook Evasion (Indirect Syscalls)
GoSneak bypasses EDR telemetry by implementing indirect syscalls through custom **AMD64 Assembly stubs**. By jumping to legitimate `syscall` instructions within `ntdll.dll`, the framework ensures the call stack originates from a trusted module, bypassing inline hooks.

### 2. Manual Memory Management
Utilizing the `unsafe` package, GoSneak circumvents the Go Garbage Collector (GC):
- **Unmanaged Allocation**: Direct interaction with `NtAllocateVirtualMemory` to reserve non-paged pool memory.
- **Permission Lifecycle**: Implements a strict **RW -> RX** transition to minimize memory artifacts.
- **Pointer Manipulation**: Uses `uintptr` arithmetic to handle payload de-obfuscation without triggering memory-copy signatures.

### 3. Cryptographic Pipeline
- **AES-256-GCM**: Authenticated decryption for payload integrity.
- **Bitwise XOR**: Lightweight de-obfuscation to disrupt static string analysis.

## 🏗️ Execution Flow
1. **Init**: Environmental sanity checks and anti-debugging triggers.
2. **Unpack**: AES-256-GCM decryption + XOR de-obfuscation in a Go-managed buffer.
3. **Allocate**: Manual allocation of non-paged pool memory via indirect syscall.
4. **Bridge**: `unsafe.Pointer` transfer from managed slice to unmanaged kernel buffer.
5. **Protect**: `VirtualProtect` transition to `PAGE_EXECUTE_READ`.
6. **Trigger**: Execution via `EnumSystemLocalesA` or similar legitimate system callbacks.

## 🛠️ Usage
1. **Clone the repo**:
   ```bash
   git clone https://github.com/scarlett-danger/GoSneak.git
   ```
2. **Build the stealth binary**:
   ```bash
   GOOS=windows GOARCH=amd64 go build -ldflags="-s -w -H=windowsgui" -o GoSneak.exe .
   ```
3. **Run the geneated binary on a Windows target**
   ```
   .\GoSneak.exe
   ```

## 🐾 Community
Want to help the Gopher sneak even better? Contributions to **GoSneak** are always welcome. Open an issue or submit a PR to help.

---
## ⚖️ LEGAL DISCLAIMER & LIMITATION OF LIABILITY

**IMPORTANT: READ CAREFULLY BEFORE PROCEEDING.**

1.  **Strictly for Legal Use**: GoSneak is developed and intended solely for authorized cybersecurity research, red-teaming exercises, and educational purposes. It is **not** intended for, nor should it be used for, any illegal or malicious activities.  Don't be a bad gopher.
2.  **No Liability**: The author(s) of GoSneak shall **not be held legally responsible** for any misuse, damage, or legal consequences resulting from the use of this software. By downloading, compiling, or running this code, you assume full responsibility for your actions.
3.  **Compliance**: Users are responsible for ensuring that their use of this framework complies with all local, state, national, and international laws and regulations. Unauthorized access to computer systems is a crime.
4.  **No Warranty**: This software is provided "as-is" without any warranty of any kind, either expressed or implied.

---

---
*Developed with 🖤 and a healthy disrespect for `reflect.Safe`*

---
