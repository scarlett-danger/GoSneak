// func indirectSyscall(ssn uint16, syscallAddr uintptr, handle uintptr, addr *uintptr, zeroBits uintptr, size *uintptr, allocType uintptr, protect uintptr) uintptr
TEXT ·indirectSyscall(SB), $0-64
    MOVQ ssn+0(FP), AX
    MOVQ syscallAddr+8(FP), R11
    MOVQ handle+16(FP), R10
    MOVQ addr+24(FP), R12
    MOVQ zeroBits+32(FP), R13
    MOVQ size+40(FP), R14
    MOVQ allocType+48(FP), R15
    // Windows syscalls expect 4+ arguments on the stack, but for simplicity
    // we are demonstrating the register transition (R10, RDX, R8, R9)
    JMP R11
    RET
