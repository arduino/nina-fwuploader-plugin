// This file is part of nina-fwuploader-plugin.
//
// Copyright (c) 2023 Arduino LLC.  All right reserved.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published
// by the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

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
