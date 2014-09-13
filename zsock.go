package goczmq

/*
#cgo !windows pkg-config: libczmq
#cgo windows CFLAGS: -I/usr/local/include
#cgo windows LDFLAGS: -L/usr/local/lib -lczmq
#include "czmq.h"
#include <stdlib.h>
#include <string.h>

int Zsock_connect(zsock_t *self, const char *format) {return zsock_connect(self, format, NULL);}
int Zsock_bind(zsock_t *self, const char *format) {return zsock_bind(self, format, NULL);}
*/
import "C"

import (
	"errors"
	"runtime"
	"unsafe"
)

// Zsock wraps the zsock_t class in CZMQ.
type Zsock struct {
	zsock_t *C.struct__zsock_t
	file    string
	line    int
	zType   Type
}

// NewZsock creates a new socket.  The caller source and
// line number are passed so CZMQ can report socket leaks
// intelligently.
func NewZsock(t Type) *Zsock {
	var z *Zsock

	_, file, line, ok := runtime.Caller(1)

	if ok {
		z = &Zsock{file: file, line: line, zType: t}
	} else {
		z = &Zsock{file: "", line: 0, zType: t}
	}

	z.zsock_t = C.zsock_new_(C.int(t), C.CString(z.file), C.size_t(z.line))
	return z
}

// Connect connects a socket to an endpoint
// returns an error if the connect failed.
func (z *Zsock) Connect(endpoint string) error {
	rc := C.Zsock_connect(z.zsock_t, C.CString(endpoint))
	if rc == C.int(-1) {
		return errors.New("failed")
	} else {
		return nil
	}
}

// Bind binds a socket to an endpoint.  On success returns
// the port number used for tcp transports, or 0 for other
// transports.  On failure returns a -1 for port, and an error.
func (z *Zsock) Bind(endpoint string) (int, error) {
	port := C.Zsock_bind(z.zsock_t, C.CString(endpoint))
	if port == C.int(-1) {
		return -1, errors.New("failed")
	} else {
		return int(port), nil
	}
}

// SendBytes sends a byte array via the socket.  For the flags
// value, use 0 for a single message, or SNDMORE if it is
// a multi-part message
func (z *Zsock) SendBytes(data []byte, flags Flag) error {
	frame := C.zframe_new(unsafe.Pointer(&data[0]), C.size_t(len(data)))
	rc := C.zframe_send(&frame, unsafe.Pointer(z.zsock_t), C.int(flags))
	if rc == C.int(-1) {
		return errors.New("failed")
	} else {
		return nil
	}
}

// SendString sends a string via the socket.  For the flags
// value, use 0 for a single message, or SNDMORE if it is
// a multi-part message
func (z *Zsock) SendString(data string, flags Flag) error {
	err := z.SendBytes([]byte(data), flags)
	return err
}

// RecvBytes reads a frame from the socket and returns it
// as a byte array,  Returns an error if the call fails.
func (z *Zsock) RecvBytes() ([]byte, error) {
	frame := C.zframe_recv(unsafe.Pointer(z.zsock_t))
	if frame == nil {
		return []byte{0}, errors.New("failed")
	}
	dataSize := C.zframe_size(frame)
	dataPtr := C.zframe_data(frame)
	b := C.GoBytes(unsafe.Pointer(dataPtr), C.int(dataSize))
	C.zframe_destroy(&frame)
	return b, nil
}

// RecvString reads a frame from the socket and returns it
// as a string,  Returns an error if the call fails.
func (z *Zsock) RecvString() (string, error) {
	b, err := z.RecvBytes()
	if err != nil {
		return "", err
	} else {
		return string(b), err
	}
}

// Destroy destroys the underlying zsock_t.
func (z *Zsock) Destroy() {
	C.zsock_destroy_(&z.zsock_t, C.CString(z.file), C.size_t(z.line))
}
