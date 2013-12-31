/* Common data file features. */
package file

import (
	"errors"
	"fmt"
	"github.com/HouzuoGuo/tiedot/gommap"
	"github.com/HouzuoGuo/tiedot/tdlog"
	"os"
)

type File struct {
	Name           string // File path and name
	UsedSize, Size uint64
	Growth         uint64
	Fh             *os.File    // File handle
	Buf            gommap.MMap // Mapped file buffer
}

// Return true if the buffer begins with twenty consecutive 0s.
func ConsecutiveTwenty0s(buf gommap.MMap) bool {
	for i := 0; i < 20; i++ {
		if buf[i] != 0 {
			return false
		}
	}
	return true
}

// Open the file, or create it if non-existing.
func Open(name string, growth uint64) (file *File, err error) {
	if growth < 1 {
		err = errors.New(fmt.Sprintf("File growth should be greater than one (opening %s)", growth, name))
	}
	file = &File{Name: name, Growth: growth}
	// Open file (get a handle) and determine its size
	if file.Fh, err = os.OpenFile(name, os.O_CREATE|os.O_RDWR, 0600); err != nil {
		return
	}
	fsize, err := file.Fh.Seek(0, os.SEEK_END)
	if err != nil {
		return
	}
	file.Size = uint64(fsize)
	// The file must have non-0 size
	if file.Size == 0 {
		file.CheckSizeAndEnsure(file.Growth)
		return
	}
	// Map the file into memory buffer
	if file.Buf, err = gommap.Map(file.Fh, gommap.RDWR, 0); err != nil {
		return
	}
	// Bi-sect file buffer to find out how much space is in-use
	for low, mid, high := uint64(0), file.Size/2, file.Size; ; {
		switch {
		case high-mid == 1:
			if ConsecutiveTwenty0s(file.Buf[mid:]) {
				if ConsecutiveTwenty0s(file.Buf[mid-1:]) {
					file.UsedSize = mid - 1
				} else {
					file.UsedSize = mid
				}
				return
			}
			file.UsedSize = high
			return
		case ConsecutiveTwenty0s(file.Buf[mid:]):
			high = mid
			mid = low + (mid-low)/2
		default:
			low = mid
			mid = mid + (high-mid)/2
		}
	}
	tdlog.Printf("%s has %d bytes out of %d bytes in-use", name, file.UsedSize, file.Size)
	return
}

// Return true only if the file still has room for more data.
func (file *File) CheckSize(more uint64) bool {
	return file.UsedSize+more <= file.Size
}

// Ensure that the file has enough room for more data. Grow the file if necessary.
func (file *File) CheckSizeAndEnsure(more uint64) {
	if file.UsedSize+more <= file.Size {
		return
	}
	// Grow the file - unmap the file, truncate and then re-map
	var err error
	if file.Buf != nil {
		if err = file.Buf.Unmap(); err != nil {
			panic(err)
		}
	}
	if err = os.Truncate(file.Name, int64(file.Size+file.Growth)); err != nil {
		panic(err)
	}
	if file.Buf, err = gommap.Map(file.Fh, gommap.RDWR, 0); err != nil {
		panic(err)
	}
	file.Size += file.Growth
	tdlog.Printf("File %s has grown %d bytes\n", file.Name, file.Growth)
	file.CheckSizeAndEnsure(more)
}

// Synchronize file buffer with underlying storage device.
func (file *File) Flush() error {
	return file.Buf.Flush()
}

// Close the file.
func (file *File) Close() error {
	if err := file.Buf.Unmap(); err != nil {
		return err
	}
	return file.Fh.Close()
}