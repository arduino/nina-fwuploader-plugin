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

#include <Arduino.h>

class ESP32BootROMClass {
  public:
    ESP32BootROMClass(HardwareSerial& hwSerial, int gpio0Pin, int resetnPin);

    int begin(unsigned long baudrate);
    void end();

    int beginFlash(uint32_t offset, uint32_t size, uint32_t chunkSize);
    int dataFlash(const void* data, uint32_t length);
    int endFlash(uint32_t reboot);

    int md5Flash(uint32_t offset, uint32_t size, uint8_t* result);

  private:
    int sync();
    int changeBaudrate(unsigned long baudrate);
    int spiAttach();

    void command(int opcode, const void* data, uint16_t length);
    int response(int opcode, unsigned long timeout, void* body = NULL);

    void writeEscapedBytes(const uint8_t* data, uint16_t length);

  private:
    HardwareSerial* _serial;
    int _gpio0Pin;
    int _resetnPin;

    uint32_t _flashSequenceNumber;
    uint32_t _chunkSize;
};

extern ESP32BootROMClass ESP32BootROM;
