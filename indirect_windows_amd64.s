// func indirectSyscall(ssn uint16, syscallAddr uintptr, handle uintptr, addr *uintptr, zeroBits uintptr, size *uintptr, allocType uintptr, protect uintptr) uintptr
TEXT ·indirectSyscall(SB), $0-64
    MOVQ ssn+0(FP), AX       // SSN into RAX
    MOVQ syscallAddr+8(FP), R11 // Syscall instruction address into R11
    MOVQ handle+16(FP), R10    // Handle into R10
    MOVQ addr+24(FP), RDX      // &baseAddress into RDX
    MOVQ zeroBits+32(FP), R8   // zeroBits into R8
    MOVQ size+40(FP), R9       // &size into R9
    
    // Setup stack for remaining 2 arguments (AllocType, Protect)
    MOVQ allocType+48(FP), R12
    MOVQ R12, 32(SP)
    MOVQ protect+56(FP), R13
    MOVQ R13, 40(SP)

    JMP R11                  // Indirect jump to syscall instruction
    RET
