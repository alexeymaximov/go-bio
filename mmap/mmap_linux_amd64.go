package mmap

import (
	"os"
	"reflect"
	"runtime"
	"syscall"
	"unsafe"
)

// errno returns a system error code.
func errno(err error) error {
	if err != nil {
		if en, ok := err.(syscall.Errno); ok && en == 0 {
			return syscall.EINVAL
		}
		return err
	}
	return syscall.EINVAL
}

// mmap wraps the system call for mmap.
func mmap(addr, length uintptr, prot, flags int, fd uintptr, offset int64) (uintptr, error) {
	if prot < 0 || flags < 0 || offset < 0 {
		return 0, syscall.EINVAL
	}
	result, _, err := syscall.Syscall6(syscall.SYS_MMAP, addr, length, uintptr(prot), uintptr(flags), fd, uintptr(offset))
	if err != 0 {
		return 0, errno(err)
	}
	return result, nil
}

// mlock wraps the system call for mlock.
func mlock(addr, length uintptr) error {
	_, _, err := syscall.Syscall(syscall.SYS_MLOCK, addr, length, 0)
	if err != 0 {
		return errno(err)
	}
	return err
}

// munlock wraps the system call for munlock.
func munlock(addr, length uintptr) error {
	_, _, err := syscall.Syscall(syscall.SYS_MUNLOCK, addr, length, 0)
	if err != 0 {
		return errno(err)
	}
	return nil
}

// msync wraps the system call for msync.
func msync(addr, length uintptr) error {
	_, _, err := syscall.Syscall(syscall.SYS_MSYNC, addr, length, syscall.MS_SYNC)
	if err != 0 {
		return errno(err)
	}
	return nil
}

// munmap wraps the system call for munmap.
func munmap(addr, length uintptr) error {
	_, _, err := syscall.Syscall(syscall.SYS_MUNMAP, addr, length, 0)
	if err != 0 {
		return errno(err)
	}
	return nil
}

// Mapping is a mapping of the file into the memory.
type Mapping struct {
	generic
	// alignedAddress specifies the start address of the the mapped memory
	// aligned by the memory page size.
	alignedAddress uintptr
	// alignedLength specifies the length of the mapped memory, in bytes,
	// aligned by the memory page size.
	alignedLength uintptr
	// locked specifies whether the mapped memory is locked.
	locked bool
}

// Open opens and returns a new mapping of the given file into the memory.
// The given file descriptor will be duplicated. It means that
// if the parent file will be closed the mapping will still be valid.
// Actual offset and length may be different than the given
// by the reason of aligning to the memory page size.
func Open(fd uintptr, offset int64, length uintptr, mode Mode, flags Flag) (*Mapping, error) {

	// Using int64 (off_t) for the offset and uintptr (size_t) for the length
	// by the reason of the compatibility.
	if offset < 0 {
		return nil, ErrBadOffset
	}
	if length > uintptr(MaxInt) {
		return nil, ErrBadLength
	}

	m := &Mapping{}
	prot := syscall.PROT_READ
	mmapFlags := syscall.MAP_SHARED
	if mode < ModeReadOnly || mode > ModeWriteCopy {
		return nil, ErrBadMode
	}
	if mode > ModeReadOnly {
		prot |= syscall.PROT_WRITE
		m.writable = true
	}
	if mode == ModeWriteCopy {
		flags = syscall.MAP_PRIVATE
	}
	if flags&FlagExecutable != 0 {
		prot |= syscall.PROT_EXEC
		m.executable = true
	}

	// The mapping address range must be aligned by the memory page size.
	pageSize := int64(os.Getpagesize())
	if pageSize < 0 {
		return nil, os.NewSyscallError("getpagesize", syscall.EINVAL)
	}
	outerOffset := offset / pageSize
	innerOffset := offset % pageSize
	// ASSERT: uintptr is of the 64-bit length on the amd64 architecture.
	m.alignedLength = uintptr(innerOffset) + length

	var err error
	m.alignedAddress, err = mmap(0, m.alignedLength, prot, mmapFlags, fd, outerOffset)
	if err != nil {
		return nil, os.NewSyscallError("mmap", err)
	}
	m.address = m.alignedAddress + uintptr(innerOffset)

	// Wrapping the mapped memory by the byte slice.
	slice := reflect.SliceHeader{}
	slice.Data = m.address
	slice.Len = int(length)
	slice.Cap = slice.Len
	m.memory = *(*[]byte)(unsafe.Pointer(&slice))

	runtime.SetFinalizer(m, (*Mapping).Close)
	return m, nil
}

// Lock locks the mapped memory pages.
// All pages that contain a part of the mapping address range
// are guaranteed to be resident in RAM when the call returns successfully.
// The pages are guaranteed to stay in RAM until later unlocked.
// It may need to increase process memory limits for operation success.
// See working set on Windows and rlimit on Linux for details.
func (m *Mapping) Lock() error {
	if m.memory == nil {
		return ErrClosed
	}
	if m.locked {
		return ErrLocked
	}
	if err := mlock(m.alignedAddress, m.alignedLength); err != nil {
		return os.NewSyscallError("mlock", err)
	}
	m.locked = true
	return nil
}

// Unlock unlocks the previously locked mapped memory pages.
func (m *Mapping) Unlock() error {
	if m.memory == nil {
		return ErrClosed
	}
	if !m.locked {
		return ErrNotLocked
	}
	if err := munlock(m.alignedAddress, m.alignedLength); err != nil {
		return os.NewSyscallError("munlock", err)
	}
	m.locked = false
	return nil
}

// Sync synchronizes the mapped memory with the underlying file.
func (m *Mapping) Sync() error {
	if m.memory == nil {
		return ErrClosed
	}
	if !m.writable {
		return ErrReadOnly
	}
	return os.NewSyscallError("msync", msync(m.alignedAddress, m.alignedLength))
}

// Close closes this mapping and frees all resources associated with it.
// Mapped memory will be synchronized with the underlying file and unlocked automatically.
// Close implements the io.Closer interface.
func (m *Mapping) Close() error {
	if m.memory == nil {
		return ErrClosed
	}
	var errs []error

	// Maybe unnecessary.
	if m.writable {
		if err := m.Sync(); err != nil {
			errs = append(errs, err)
		}
	}
	if m.locked {
		if err := m.Unlock(); err != nil {
			errs = append(errs, err)
		}
	}

	if err := munmap(m.alignedAddress, m.alignedLength); err != nil {
		errs = append(errs, os.NewSyscallError("munmap", err))
	}
	*m = Mapping{}
	runtime.SetFinalizer(m, nil)
	if len(errs) > 0 {
		return errs[0]
	}
	return nil
}
