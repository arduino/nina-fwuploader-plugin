package main

import (
	"bytes"
	"crypto/md5"
	"encoding/binary"
	"errors"
	"fmt"
	"time"

	helper "github.com/arduino/fwuploader-plugin-helper"
	"go.bug.st/serial"
)

type commandData struct {
	Command byte
	Address uint32
	Value   uint32
	Payload []byte
}

type flasher struct {
	port             serial.Port
	payloadSize      int
	progressCallback func(progress int)
}

func defaultProgressCallBack(feedback *helper.PluginFeedback) func(progress int) {
	return func(progress int) { fmt.Fprintf(feedback.Out(), "Flashing progress: %d%%\r", progress) }
}

func newFlasher(portAddress string, progressCallback func(progress int)) (*flasher, error) {
	port, err := serialOpen(portAddress)
	if err != nil {
		return nil, err
	}

	// wait 2 seconds to ensure the serial port is ready
	time.Sleep(2 * time.Second)
	reboot(port)
	time.Sleep(2 * time.Second)

	f := &flasher{port: port, progressCallback: progressCallback}

	payloadSize, err := f.getMaximumPayloadSize()
	if err != nil {
		return nil, err
	}
	if payloadSize < 1024 {
		return nil, fmt.Errorf("programmer reports %d as maximum payload size (1024 is needed)", payloadSize)
	}
	f.payloadSize = int(payloadSize)

	return f, nil
}

// flashChunk flashes a chunk of data
func (f *flasher) flashChunk(offset int, buffer []byte) error {
	chunkSize := int(f.payloadSize)
	bufferLength := len(buffer)

	if err := f.erase(uint32(offset), uint32(bufferLength)); err != nil {
		return err
	}

	for i := 0; i < bufferLength; i += chunkSize {
		if f.progressCallback != nil {
			progress := (i * 100) / bufferLength
			f.progressCallback(progress)
		}
		start := i
		end := i + chunkSize
		if end > bufferLength {
			end = bufferLength
		}
		if err := f.write(uint32(offset+i), buffer[start:end]); err != nil {
			return err
		}
	}

	return nil
}

// getMaximumPayloadSize asks the board the maximum payload size
func (f *flasher) getMaximumPayloadSize() (uint16, error) {
	// "MAX_PAYLOAD_SIZE" command
	err := f.sendCommand(commandData{
		Command: 0x50,
		Address: 0,
		Value:   0,
		Payload: nil,
	})
	if err != nil {
		return 0, err
	}

	// Receive response
	res := make([]byte, 2)
	if err := f.serialFillBuffer(res); err != nil {
		return 0, err
	}
	return (uint16(res[0]) << 8) + uint16(res[1]), nil
}

// serialFillBuffer fills buffer with data read from the serial port
func (f *flasher) serialFillBuffer(buffer []byte) error {
	read := 0
	for read < len(buffer) {
		n, err := f.port.Read(buffer[read:])
		if err != nil {
			return err
		}
		if n == 0 {
			return errors.New("serial port closed unexpectedly")
		}
		read += n
	}
	return nil
}

// sendCommand sends the data over serial port to connected board
func (f *flasher) sendCommand(data commandData) error {
	buf := &bytes.Buffer{}
	if err := binary.Write(buf, binary.BigEndian, data.Command); err != nil {
		return fmt.Errorf("writing command: %s", err)
	}
	if err := binary.Write(buf, binary.BigEndian, data.Address); err != nil {
		return fmt.Errorf("writing address: %s", err)
	}
	if err := binary.Write(buf, binary.BigEndian, data.Value); err != nil {
		return fmt.Errorf("writing value: %s", err)
	}
	if err := binary.Write(buf, binary.BigEndian, uint16(len(data.Payload))); err != nil {
		return fmt.Errorf("writing payload length: %s", err)
	}
	if data.Payload != nil {
		buf.Write(data.Payload)
	}
	bufferData := buf.Bytes()
	for {
		sent, err := f.port.Write(bufferData)
		if err != nil {
			return fmt.Errorf("writing data: %s", err)
		}
		if sent == len(bufferData) {
			break
		}
		bufferData = bufferData[sent:]
	}
	return nil
}

// read a block of flash memory
func (f *flasher) read(address uint32, length uint32) ([]byte, error) {
	// "FLASH_READ" command
	err := f.sendCommand(commandData{
		Command: 0x01,
		Address: address,
		Value:   length,
		Payload: nil,
	})
	if err != nil {
		return nil, err
	}

	// Receive response
	result := make([]byte, length)
	if err := f.serialFillBuffer(result); err != nil {
		return nil, err
	}
	ack := make([]byte, 2)
	if err := f.serialFillBuffer(ack); err != nil {
		return nil, err
	}
	if string(ack) != "OK" {
		return nil, fmt.Errorf("missing ack on read: %s, result: %s", ack, result)
	}
	return result, nil
}

// write a block of flash memory
func (f *flasher) write(address uint32, buffer []byte) error {
	// "FLASH_WRITE" command
	err := f.sendCommand(commandData{
		Command: 0x02,
		Address: address,
		Value:   0,
		Payload: buffer,
	})
	if err != nil {
		return err
	}

	// wait acknowledge
	ack := make([]byte, 2)
	if err := f.serialFillBuffer(ack); err != nil {
		return err
	}
	if string(ack) != "OK" {
		return fmt.Errorf("missing ack on write: %s", ack)
	}
	return nil
}

// erase a block of flash memory
func (f *flasher) erase(address uint32, length uint32) error {
	// "FLASH_ERASE" command
	err := f.sendCommand(commandData{
		Command: 0x03,
		Address: address,
		Value:   length,
		Payload: nil,
	})
	if err != nil {
		return err
	}

	// wait acknowledge
	ack := make([]byte, 2)
	if err := f.serialFillBuffer(ack); err != nil {
		return err
	}
	if string(ack) != "OK" {
		return fmt.Errorf("missing ack on erase: %s", ack)
	}
	return nil
}

func (f *flasher) md5sum(data []byte) error {
	hasher := md5.New()
	hasher.Write(data)

	// Get md5sum
	if err := f.sendCommand(commandData{
		Command: 0x04,
		Address: 0,
		Value:   uint32(len(data)),
		Payload: nil,
	}); err != nil {
		return err
	}

	// Wait acknowledge
	ack := make([]byte, 2)
	if err := f.serialFillBuffer(ack); err != nil {
		return err
	}
	if string(ack) != "OK" {
		return fmt.Errorf("missing ack on md5sum: %s", ack)
	}

	// Wait md5
	md5sumfromdevice := make([]byte, 16)
	if err := f.serialFillBuffer(md5sumfromdevice); err != nil {
		return err
	}

	md5sum := hasher.Sum(nil)

	for i := 0; i < 16; i++ {
		if md5sum[i] != md5sumfromdevice[i] {
			return fmt.Errorf("MD5sum failed")
		}
	}

	return nil
}
