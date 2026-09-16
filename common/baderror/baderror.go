package baderror

import (
	"context"
	"errors"
	"io"
	"net"
	"strings"
)

func Contains(err error, msgList ...string) bool {
	for _, msg := range msgList {
		if strings.Contains(err.Error(), msg) {
			return true
		}
	}
	return false
}

func WrapH2(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, io.ErrUnexpectedEOF) {
		return io.EOF
	}
	if Contains(err, "client disconnected", "body closed by handler", "response body closed", "; CANCEL") {
		return net.ErrClosed
	}
	if Contains(err, "stream error: ") {
		// an HTTP/2 connection reading a bare StreamError from its conn takes it as its own and spins forever
		return &h2StreamError{err}
	}
	return err
}

type h2StreamError struct {
	error
}

func (e *h2StreamError) Unwrap() error {
	return e.error
}

func WrapGRPC(err error) error {
	// grpc uses stupid internal error types
	if err == nil {
		return nil
	}
	if Contains(err, "EOF") {
		return io.EOF
	}
	if Contains(err, "Canceled") {
		return context.Canceled
	}
	if Contains(err,
		"the client connection is closing",
		"server closed the stream without sending trailers") {
		return net.ErrClosed
	}
	return err
}

// Deprecated: use qtls.WrapError
func WrapQUIC(err error) error {
	if err == nil {
		return nil
	}
	if Contains(err,
		"canceled by remote with error code 0",
		"canceled by local with error code 0",
	) {
		return net.ErrClosed
	}
	return err
}
