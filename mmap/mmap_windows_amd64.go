package mmap

import (
	"math"
	"os"
	"runtime"
	"syscall"
	"unsafe"
)

// Mapping is a mapping of the file into the memory.
type Mapping struct {
	generic
	// hProcess specifies the descriptor of the current process.
	hProcess syscall.Handle
	// hFile specifies the descriptor of the mapped file.
	hFile syscall.Handle
	// hMapping specifies the descriptor of the mapping object provided by the operation system.
	hMapping syscall.Handle
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
	prot := uint32(syscall.PAGE_READONLY)
	access := uint32(syscall.FILE_MAP_READ)
	switch mode {
	case ModeReadOnly:
		// NOOP
	case ModeReadWrite:
		prot = syscall.PAGE_READWRITE
		access = syscall.FILE_MAP_WRITE
		m.writable = true
	case ModeWriteCopy:
		prot = syscall.PAGE_WRITECOPY
		access = syscall.FILE_MAP_COPY
		m.writable = true
	default:
		return nil, ErrBadMode
	}
	if flags&FlagExecutable != 0 {
		prot <<= 4
		access |= syscall.FILE_MAP_EXECUTE
		m.executable = true
	}

	// The separate file handle is needed to avoid errors on the mapped file external closing.
	var err error
	m.hProcess, err = syscall.GetCurrentProcess()
	if err != nil {
		return nil, os.NewSyscallError("GetCurrentProcess", err)
	}
	err = syscall.DuplicateHandle(
		m.hProcess, syscall.Handle(fd),
		m.hProcess, &m.hFile,
		0, true, syscall.DUPLICATE_SAME_ACCESS,
	)
	if err != nil {
		return nil, os.NewSyscallError("DuplicateHandle", err)
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

	maxSize := uint64(outerOffset) + uint64(m.alignedLength)
	maxSizeHigh := uint32(maxSize >> 32)
	maxSizeLow := uint32(maxSize & uint64(math.MaxUint32))
	m.hMapping, err = syscall.CreateFileMapping(m.hFile, nil, prot, maxSizeHigh, maxSizeLow, nil)
	if err != nil {
		return nil, os.NewSyscallError("CreateFileMapping", err)
	}
	fileOffset := uint64(outerOffset)
	fileOffsetHigh := uint32(fileOffset >> 32)
	fileOffsetLow := uint32(fileOffset & uint64(math.MaxUint32))
	m.alignedAddress, err = syscall.MapViewOfFile(
		m.hMapping, access,
		fileOffsetHigh, fileOffsetLow, m.alignedLength,
	)
	if err != nil {
		return nil, os.NewSyscallError("MapViewOfFile", err)
	}
	m.address = m.alignedAddress + uintptr(innerOffset)

	// Wrapping the mapped memory by the byte slice.
	var slice struct {
		ptr uintptr
		len int
		cap int
	}
	slice.ptr = m.address
	slice.len = int(length)
	slice.cap = slice.len
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
	if err := syscall.VirtualLock(m.alignedAddress, m.alignedLength); err != nil {
		return os.NewSyscallError("VirtualLock", err)
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
	if err := syscall.VirtualUnlock(m.alignedAddress, m.alignedLength); err != nil {
		return os.NewSyscallError("VirtualUnlock", err)
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
	if err := syscall.FlushViewOfFile(m.alignedAddress, m.alignedLength); err != nil {
		return os.NewSyscallError("FlushViewOfFile", err)
	}
	if err := syscall.FlushFileBuffers(m.hFile); err != nil {
		return os.NewSyscallError("FlushFileBuffers", err)
	}
	return nil
}

// Close closes this mapping and frees all resources associated with it.
// Mapped memory will be synchronized with the underlying file and unlocked automatically.
// Close implements the io.Closer interface.
func (m *Mapping) Close() error {
	if m.memory == nil {
		return ErrClosed
	}
	var errs []error
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
	if err := syscall.UnmapViewOfFile(m.alignedAddress); err != nil {
		errs = append(errs, os.NewSyscallError("UnmapViewOfFile", err))
	}
	if err := syscall.CloseHandle(m.hMapping); err != nil {
		errs = append(errs, os.NewSyscallError("CloseHandle", err))
	}
	if err := syscall.CloseHandle(m.hFile); err != nil {
		errs = append(errs, os.NewSyscallError("CloseHandle", err))
	}
	*m = Mapping{}
	runtime.SetFinalizer(m, nil)
	if len(errs) > 0 {
		return errs[0]
	}
	return nil
}
