package main

// Very simple utility which signals an event. Used to signal a process
// on on Windows to either dump its stacks or it's debugger event.
// Works across dockerd.exe; containerd.exe; containerd-shim-runhcs-v1.exe
// Flags: -pid=daemonpid [-debugger]

// go build -o signal-event.exe

import (
	"flag"
	"fmt"
	"syscall"
	"unsafe"
)

const EVENT_MODIFY_STATUS = 0x0002

var (
	modkernel32    = syscall.NewLazyDLL("kernel32.dll")
	procOpenEvent  = modkernel32.NewProc("OpenEventW")
	procPulseEvent = modkernel32.NewProc("PulseEvent")
)

func OpenEvent(desiredAccess uint32, inheritHandle bool, name string) (handle syscall.Handle, err error) {
	namep, _ := syscall.UTF16PtrFromString(name)
	var _p2 uint32 = 0
	if inheritHandle {
		_p2 = 1
	}
	r0, _, e1 := procOpenEvent.Call(uintptr(desiredAccess), uintptr(_p2), uintptr(unsafe.Pointer(namep)))
	use(unsafe.Pointer(namep))
	handle = syscall.Handle(r0)
	if handle == syscall.InvalidHandle {
		err = e1
	}
	return
}

func PulseEvent(handle syscall.Handle) (err error) {
	r0, _, _ := procPulseEvent.Call(uintptr(handle))
	if r0 != 0 {
		err = syscall.Errno(r0)
	}
	return
}

func main() {
	var (
		pid      int
		debugger bool
	)
	flag.BoolVar(&debugger, "debugger", false, "Signal a debugger event rather than stackdump.")
	flag.IntVar(&pid, "pid", -1, "PID of process to signal")
	flag.Parse()
	if pid == -1 {
		fmt.Println("Error: pid must be supplied")
		return
	}
	key := "stackdump"
	if debugger {
		key = "debugger"
	}

	ev := fmt.Sprintf("Global\\%s-%s", key, fmt.Sprint(pid))
	h2, _ := OpenEvent(EVENT_MODIFY_STATUS, false, ev)
	if h2 == 0 {
		fmt.Printf("Could not open event. Check PID %d is correct and is running.\n", pid)
		return
	}
	PulseEvent(h2)
	fmt.Println("Signalled successfully.")
}

var temp unsafe.Pointer

func use(p unsafe.Pointer) {
	temp = p
}
