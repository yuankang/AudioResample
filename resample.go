package main

/*
#cgo LDFLAGS: -lsoxr

#include <stdlib.h>
#include "soxr.h"
*/
import "C"
import (
	"errors"
	"runtime"
	"unsafe"
)

const (
	Quick     = 0 // Quick cubic interpolation
	LowQ      = 1 // LowQ 16-bit with larger rolloff
	MediumQ   = 2 // MediumQ 16-bit with medium rolloff
	HighQ     = 4 // High quality
	VeryHighQ = 6 // Very high quality

	F32 = 0 // 32-bit floating point PCM
	F64 = 1 // 64-bit floating point PCM
	I32 = 2 // 32-bit signed linear PCM
	I16 = 3 // 16-bit signed linear PCM

	byteLen = 8
)

type Resampler struct {
	resampler C.soxr_t
	inRate    float64 // input sample rate
	outRate   float64 // output sample rate
	channels  int     // number of input channels
	frameSize int     // frame size in bytes
}

var threads int

func init() {
	threads = runtime.NumCPU()
}

func New(inputRate, outputRate float64, channels, format, quality int) (*Resampler, error) {
	var err error
	var size int
	if inputRate <= 0 || outputRate <= 0 {
		return nil, errors.New("Invalid input or output sampling rates")
	}
	if channels == 0 {
		return nil, errors.New("Invalid channels number")
	}
	if quality < 0 || quality > 6 {
		return nil, errors.New("Invalid quality setting")
	}
	switch format {
	case F64:
		size = 64 / byteLen
	case F32, I32:
		size = 32 / byteLen
	case I16:
		size = 16 / byteLen
	default:
		return nil, errors.New("Invalid format setting")
	}
	var soxr C.soxr_t
	var soxErr C.soxr_error_t
	// Setup soxr and create a stream resampler
	ioSpec := C.soxr_io_spec(C.soxr_datatype_t(format), C.soxr_datatype_t(format))
	qSpec := C.soxr_quality_spec(C.ulong(quality), 0)
	runtimeSpec := C.soxr_runtime_spec(C.uint(threads))
	soxr = C.soxr_create(C.double(inputRate), C.double(outputRate), C.uint(channels), &soxErr, &ioSpec, &qSpec, &runtimeSpec)
	if C.GoString(soxErr) != "" && C.GoString(soxErr) != "0" {
		err = errors.New(C.GoString(soxErr))
		C.free(unsafe.Pointer(soxErr))
		return nil, err
	}

	r := Resampler{
		resampler: soxr,
		inRate:    inputRate,
		outRate:   outputRate,
		channels:  channels,
		frameSize: size,
	}
	C.free(unsafe.Pointer(soxErr))
	return &r, err
}

func (r *Resampler) Close() (err error) {
	if r.resampler == nil {
		return errors.New("soxr resampler is nil")
	}
	C.soxr_delete(r.resampler)
	r.resampler = nil
	return
}

func (r *Resampler) Write(p []byte) (d []byte, err error) {
	if r.resampler == nil {
		err = errors.New("soxr resampler is nil")
		return
	}
	if len(p) == 0 {
		return
	}
	if fragment := len(p) % (r.frameSize * r.channels); fragment != 0 {
		p = p[:len(p)-fragment]
	}
	framesIn := len(p) / r.frameSize / r.channels
	if framesIn == 0 {
		err = errors.New("Incomplete input frame data")
		return
	}
	framesOut := int(float64(framesIn) * (r.outRate / r.inRate))
	if framesOut == 0 {
		err = errors.New("Not enough input to generate output")
		return
	}
	dataIn := C.CBytes(p)
	dataOut := C.malloc(C.size_t(framesOut * r.channels * r.frameSize))
	var soxErr C.soxr_error_t
	var read, done C.size_t = 0, 0
	for int(done) < framesOut {
		soxErr = C.soxr_process(r.resampler, C.soxr_in_t(dataIn), C.size_t(framesIn), &read, C.soxr_out_t(dataOut), C.size_t(framesOut), &done)
		if C.GoString(soxErr) != "" && C.GoString(soxErr) != "0" {
			err = errors.New(C.GoString(soxErr))
			goto cleanup
		}
		if int(read) == framesIn && int(done) < framesOut {
			// Indicate end of input to the resampler
			var d C.size_t = 0
			soxErr = C.soxr_process(r.resampler, C.soxr_in_t(nil), C.size_t(0), nil, C.soxr_out_t(dataOut), C.size_t(framesOut), &d)
			if C.GoString(soxErr) != "" && C.GoString(soxErr) != "0" {
				err = errors.New(C.GoString(soxErr))
				goto cleanup
			}
			done += d
			break
		}
	}
	//log.Println(int(done) * r.channels * r.frameSize)
	d = C.GoBytes(dataOut, C.int(int(done)*r.channels*r.frameSize))

cleanup:
	C.free(dataIn)
	C.free(dataOut)
	C.free(unsafe.Pointer(soxErr))
	return
}
