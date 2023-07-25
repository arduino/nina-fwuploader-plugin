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

// To generate the binaries run:
// arduino-cli compile -e --profile <profile_name>

#include <SPI.h>
#include <WiFiNINA.h>
#ifdef ARDUINO_SAMD_MKRVIDOR4000
#include <VidorPeripherals.h>

unsigned long baud = 119400;
#else
unsigned long baud = 115200;
#endif

int rts = -1;
int dtr = -1;

void setup(){
  Serial.begin(baud);
  SerialNina.begin(baud);

#ifdef ARDUINO_SAMD_MKRVIDOR4000
  FPGA.begin();
#endif

#ifdef ARDUINO_SAMD_MKRVIDOR4000
  FPGA.pinMode(FPGA_NINA_GPIO0, OUTPUT);
  FPGA.pinMode(FPGA_SPIWIFI_RESET, OUTPUT);
#else
  pinMode(NINA_GPIO0, OUTPUT);
  pinMode(NINA_RESETN, OUTPUT);
#endif
  while(true) {
    if (Serial.available()){
      char choice = Serial.read();
      switch (choice){
        case 'r':
          restartPassthrough();
          return;
        case 'v':
          version();
          return;
        default:
          return;
      }
    }
  }
}

void restartPassthrough(){
#ifdef ARDUINO_AVR_UNO_WIFI_REV2
  // manually put the NINA in upload mode
  digitalWrite(NINA_GPIO0, LOW);
  digitalWrite(NINA_RESETN, LOW);
  delay(100);
  digitalWrite(NINA_RESETN, HIGH);
  delay(100);
  digitalWrite(NINA_RESETN, LOW);
#endif
  while (true) {
#ifndef ARDUINO_AVR_UNO_WIFI_REV2
    if (rts != Serial.rts()) {
#ifdef ARDUINO_SAMD_MKRVIDOR4000
      FPGA.digitalWrite(FPGA_SPIWIFI_RESET, (Serial.rts() == 1) ? LOW : HIGH);
#elif defined(ARDUINO_SAMD_NANO_33_IOT)
      digitalWrite(NINA_RESETN, Serial.rts() ? LOW : HIGH);
#else
      digitalWrite(NINA_RESETN, Serial.rts());
#endif
      rts = Serial.rts();
    }

    if (dtr != Serial.dtr()) {
#ifdef ARDUINO_SAMD_MKRVIDOR4000
      FPGA.digitalWrite(FPGA_NINA_GPIO0, (Serial.dtr() == 1) ? HIGH : LOW);
#else
      digitalWrite(NINA_GPIO0, (Serial.dtr() == 0) ? HIGH : LOW);
#endif
      dtr = Serial.dtr();
    }
#endif

    if (Serial.available()){
      SerialNina.write(Serial.read());
    }

    if (SerialNina.available()) {
      Serial.write(SerialNina.read());
    }

#ifndef ARDUINO_AVR_UNO_WIFI_REV2
    // check if the USB virtual serial wants a new baud rate
    if (Serial.baud() != baud) {
      rts = -1;
      dtr = -1;

      baud = Serial.baud();
#ifndef ARDUINO_SAMD_MKRVIDOR4000
      SerialNina.begin(baud);
#endif
    }
#endif
  }
}

void version(){
  if (WiFi.status() == WL_NO_MODULE) {
    Serial.println("99.99.99");
    return;
  }

  // Print firmware version on the module
  String fv = WiFi.firmwareVersion();
  Serial.println(fv);
}

void loop() {
}
