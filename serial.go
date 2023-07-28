package main

import (
	"fmt"
	"time"

	"go.bug.st/serial"
)

func serialOpen(portAddress string) (serial.Port, error) {
	port, err := serial.Open(portAddress, &serial.Mode{
		BaudRate: 1000000,
		Parity:   serial.NoParity,
		DataBits: 8,
		StopBits: serial.OneStopBit,
	})
	if err != nil {
		return nil, err
	}
	if err := port.SetReadTimeout(30 * time.Second); err != nil {
		return nil, err
	}
	return port, nil
}

func reboot(port serial.Port) error {
	if _, err := port.Write([]byte("r")); err != nil {
		return fmt.Errorf("write to serial port: %v", err)
	}
	return nil
}

func getVersion(port serial.Port) error {
	if _, err := port.Write([]byte("v")); err != nil {
		return fmt.Errorf("write to serial port: %v", err)
	}
	return nil
}
