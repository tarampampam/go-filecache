// File implementation for using as a cache entry.
package file

import (
	"bytes"
	"crypto/sha1" //nolint:gosec
	"encoding/binary"
	"errors"
	"fmt"
	"hash"
	"io"
	"os"
	"time"
)

// Read/write buffer size in bytes
const rwBufferSize byte = 32

type (
	// File signature
	FSignature []byte

	// File field offset and length
	offset uint8
	length uint8

	// File field for storing "File signature" (special flag for osFile identification among other files)
	ffSignature struct {
		offset
		length
	}

	// File field for storing "Expires At" label (in unix timestamp format with milliseconds)
	ffExpiresAtUnixMs struct {
		offset
		length
	}

	// File field for storing data "hash sum" (in SHA1 format)
	ffDataSha1 struct {
		offset
		length
	}

	// Field for useful data
	ffData struct {
		offset
	}

	// Cache osFile representation (all offsets must be set manually on instance creation action)
	File struct {
		ffSignature
		ffExpiresAtUnixMs
		ffDataSha1
		ffData
		Signature FSignature
		osFile    *os.File  // osFile on filesystem
		hashing   hash.Hash // SHA1 "generator" (required for hash sum calculation)
	}
)

var DefaultSignature = FSignature("#/CACHE ") // 35, 47, 67, 65, 67, 72, 69, 32

// newFile creates new osFile instance.
func newFile(osFile *os.File, signature FSignature) *File {
	// setup default osFile type bytes slice
	if signature == nil {
		signature = DefaultSignature
	}

	// File block offsets are below:
	// +----------------+-----------------------+-----------------+------------+
	// | Signature 0..7 |    Meta Data 8..63    | DataSHA1 64..83 | Data 84..n |
	// +----------------+-----------------------+-----------------+------------+
	// |                | ExpiresAtUnixMs 8..15 |                 |            |
	// +----------------+-----------------------+-----------------+------------+
	// |                |    RESERVED 16..63    |                 |            |
	// +----------------+-----------------------+-----------------+------------+
	return &File{
		ffSignature: ffSignature{
			offset: 0,
			length: 8,
		},
		ffExpiresAtUnixMs: ffExpiresAtUnixMs{
			offset: 8,
			length: 8,
		},
		ffDataSha1: ffDataSha1{
			offset: 64,
			length: 20,
		},
		ffData: ffData{
			offset: 84,
		},
		Signature: signature,
		osFile:    osFile,
		hashing:   sha1.New(), //nolint:gosec
	}
}

// Create or truncates the named osFile. If the osFile already exists, it will be truncated. If the osFile does not exist,
// it is created with passed mode (permissions).
// signature can be omitted (nil) - in this case will be used default osFile signature.
// Important: osFile with signature and data hashsum will be created immediately.
func Create(name string, perm os.FileMode, signature FSignature) (*File, error) {
	f, openErr := os.OpenFile(name, os.O_RDWR|os.O_CREATE|os.O_TRUNC, perm)
	if openErr != nil {
		return nil, openErr
	}

	file := newFile(f, signature)

	// write osFile signature
	if err := file.setSignature(file.Signature); err != nil {
		return nil, err
	}

	// requires for hashsum init
	if err := file.SetData(bytes.NewBuffer([]byte{})); err != nil {
		return nil, err
	}

	return file, nil
}

// Open the named osFile for reading and writing. If successful, methods on the returned osFile can be used for
// reading and writing. If there is an error, it will be of type *os.PathError.
// signature can be omitted (nil) - in this case will be used default osFile signature.
func Open(name string, perm os.FileMode, signature FSignature) (*File, error) {
	return open(name, os.O_RDWR, perm, signature)
}

// OpenRead opens the named osFile for reading. If successful, methods on the returned osFile can be used for reading; the
// associated osFile descriptor has mode O_RDONLY. If there is an error, it will be of type *os.PathError.
// signature can be omitted (nil) - in this case will be used default osFile signature.
func OpenRead(name string, signature FSignature) (*File, error) {
	return open(name, os.O_RDONLY, 0, signature)
}

func open(name string, flag int, perm os.FileMode, signature FSignature) (*File, error) {
	f, err := os.OpenFile(name, flag, perm)
	if err != nil {
		return nil, err
	}

	return newFile(f, signature), nil
}

// Name returns the name of the osFile as presented to Open.
func (file *File) Name() string { return file.osFile.Name() }

// Close the File, rendering it unusable for I/O. On files that support SetDeadline, any pending I/O operations
// will be canceled and return immediately with an error.
// Close will return an error if it has already been called.
func (file *File) Close() error {
	return file.osFile.Close()
}

// SignatureMatched checks for osFile signature matching. Signature should be set on osFile creation. This function can
// helps you to detect files that created by current package.
func (file *File) SignatureMatched() (bool, error) {
	fType, err := file.getSignature()
	if err != nil {
		return false, err
	}

	return bytes.Equal(*fType, file.Signature), nil
}

// GetSignature of current osFile signature as a typed slice of a bytes.
func (file *File) GetSignature() (*FSignature, error) { return file.getSignature() }

// getSignature of current osFile signature as a typed slice of a bytes.
func (file *File) getSignature() (*FSignature, error) {
	buf := make(FSignature, file.ffSignature.length)

	if n, err := file.osFile.ReadAt(buf, int64(file.ffSignature.offset)); err != nil && err != io.EOF {
		return nil, err
	} else if l := len(buf); n != l {
		// limit length for too small reading results
		buf = buf[0:n]
	}

	return &buf, nil
}

// setSignature allows to use only bytes slice of signature with length defined in osFile structure.
func (file *File) setSignature(signature FSignature) error {
	if l := len(signature); l != int(file.ffSignature.length) {
		return fmt.Errorf("wrong signature length: required length: %d, passed: %d", file.ffSignature.length, l)
	}

	if n, err := file.osFile.WriteAt(signature, int64(file.ffSignature.offset)); err != nil {
		return err
	} else if n != len(signature) {
		return errors.New("wrong wrote bytes length")
	}

	return nil
}

// GetExpiresAt for current osFile (with milliseconds).
func (file *File) GetExpiresAt() (time.Time, error) {
	ms, err := file.getExpiresAtUnixMs()

	// check for "value was set?"
	if ms == 0 && err == nil {
		err = errors.New("value was not set")
	}

	return time.Unix(0, int64(ms*uint64(time.Millisecond))), err
}

// getExpiresAtUnixMs returns unsigned integer value with ExpiresAt in UNIX timestamp format in milliseconds.
func (file *File) getExpiresAtUnixMs() (uint64, error) {
	buf := make([]byte, file.ffExpiresAtUnixMs.length)

	if _, err := file.osFile.ReadAt(buf, int64(file.ffExpiresAtUnixMs.offset)); err != nil && err != io.EOF {
		return 0, err
	}

	return binary.LittleEndian.Uint64(buf), nil
}

// SetExpiresAt sets the expiring value.
func (file *File) SetExpiresAt(t time.Time) error {
	return file.setExpiresAtUnixMs(uint64(t.UnixNano() / int64(time.Millisecond)))
}

// setExpiresAtUnixMs sets the expiring time in milliseconds in osFile content.
func (file *File) setExpiresAtUnixMs(ts uint64) error {
	buf := make([]byte, file.ffExpiresAtUnixMs.length)

	// pack unsigned integer into slice of bytes
	binary.LittleEndian.PutUint64(buf, ts)

	if n, err := file.osFile.WriteAt(buf, int64(file.ffExpiresAtUnixMs.offset)); err != nil {
		return err
	} else if n != len(buf) {
		return errors.New("wrong wrote bytes length")
	}

	return nil
}

// setDataSHA1 sets data hashsum as s slice ob bytes. Hash length must be correct.
func (file *File) setDataSHA1(h []byte) error {
	if l := len(h); l != int(file.ffDataSha1.length) {
		return fmt.Errorf("wrong hash length: required length: %d, passed: %d", file.ffDataSha1.length, l)
	}

	if n, err := file.osFile.WriteAt(h, int64(file.ffDataSha1.offset)); err != nil {
		return err
	} else if n != len(h) {
		return errors.New("wrong wrote bytes length")
	}

	return nil
}

// GetDataHash returns osFile data hash.
func (file *File) GetDataHash() ([]byte, error) { return file.getDataSHA1() }

// getDataSHA1 returns osFile data hash.
func (file *File) getDataSHA1() ([]byte, error) {
	buf := make([]byte, file.ffDataSha1.length)

	if _, err := file.osFile.ReadAt(buf, int64(file.ffDataSha1.offset)); err != nil && err != io.EOF {
		return buf, err
	}

	return buf, nil
}

// SetData sets the osFile data (content will be read from the passed reader instance).
func (file *File) SetData(in io.Reader) error { return file.setData(in) }

// setData sets the osFile data (content will be read from the passed reader instance).
func (file *File) setData(in io.Reader) error {
	buf := make([]byte, rwBufferSize)
	off := int64(file.ffData.offset)
	file.hashing.Reset()

	for {
		// read part of input data
		n, err := in.Read(buf)
		if err != nil {
			if err != io.EOF {
				return err
			}
			break
		}

		// limit length for too small reading results
		if l := len(buf); n != l {
			buf = buf[0:n]
		}

		// write content into required position
		wroteBytes, writeErr := file.osFile.WriteAt(buf, off)
		if writeErr != nil {
			return writeErr
		}
		// write into "hashing" too for hash sum calculation
		if _, err := file.hashing.Write(buf); err != nil {
			return err
		}

		// move offset
		off += int64(wroteBytes)
	}

	if err := file.setDataSHA1(file.hashing.Sum(nil)); err != nil {
		return err
	}

	return nil
}

// GetData read osFile data and write it to the writer.
func (file *File) GetData(out io.Writer) error { return file.getData(out) }

// getData read osFile data and write it to the writer.
func (file *File) getData(out io.Writer) error {
	buf := make([]byte, rwBufferSize)
	off := uint64(file.ffData.offset)
	file.hashing.Reset()

	for {
		// read part of useful data
		n, readErr := file.osFile.ReadAt(buf, int64(off))

		// Ignore EOF here (will be checked later). In any another case - we will return an error immediately
		if readErr != nil && readErr != io.EOF {
			return readErr
		}

		// limit length for too small reading results
		if l := len(buf); n != l {
			buf = buf[0:n]
		}

		// write content into out writer
		wroteBytes, writeErr := out.Write(buf)
		if writeErr != nil {
			return writeErr
		}

		// write into "hashing" too for hash sum calculation
		if _, err := file.hashing.Write(buf); err != nil {
			return err
		}

		// move offset
		off += uint64(wroteBytes)

		if readErr == io.EOF {
			break
		}
	}

	// calculate just read data hash
	dataHash := file.hashing.Sum(nil)

	// get existing hash
	existsHash, hashErr := file.getDataSHA1()
	if hashErr != nil {
		return hashErr
	}

	// if hashes mismatched - data was broken
	if !bytes.Equal(dataHash, existsHash) {
		return fmt.Errorf("data hashes mismatched. required: %v, current: %v", existsHash, dataHash)
	}

	return nil
}
