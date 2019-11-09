package mmap

import "os"

// OpenFile prepares a file, calls the initializer if file was just created
// and returns a new mapping of the prepared file into the memory.
func OpenFile(name string, perm os.FileMode, size uintptr, flags Flag, init func(m *Mapping) error) (*Mapping, error) {
	m, created, err := func() (*Mapping, bool, error) {
		created := false
		if _, err := os.Stat(name); err != nil && os.IsNotExist(err) {
			created = true
		}
		f, err := os.OpenFile(name, os.O_CREATE|os.O_RDWR, perm)
		if err != nil {
			return nil, false, err
		}
		defer func() {
			if f != nil {
				_ = f.Close()
			}
		}()
		onFailure := func() {
			_ = f.Close()
			f = nil
			if created {
				_ = os.Remove(name)
			}
		}
		if err := f.Truncate(int64(size)); err != nil {
			onFailure()
			return nil, false, err
		}
		m, err := Open(f.Fd(), 0, size, ModeReadWrite, flags)
		if err != nil {
			onFailure()
			return nil, false, err
		}
		return m, created, nil
	}()
	if err != nil {
		return nil, err
	}
	if created && init != nil {
		if err := init(m); err != nil {
			_ = m.Close()
			_ = os.Remove(name)
			return nil, err
		}
	}
	return m, nil
}
