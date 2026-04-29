# GoSneak
A Go-native offensive security framework for User-Mode Hook Evasion and Manual Memory Management. Implements Indirect Syscalls via custom ASM stubs to bypass EDR telemetry and utilizes RWX-transitioning memory pages outside of the Go Runtime's Garbage Collector (GC) heap.
