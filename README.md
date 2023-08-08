# nina-fwuploader-plugin

[![Check Go status](https://github.com/arduino/nina-fwuploader-plugin/actions/workflows/check-go-task.yml/badge.svg)](https://github.com/arduino/nina-fwuploader-plugin/actions/workflows/check-go-task.yml)

The `nina-fwuploader-plugin` is a core component of the [arduino-fwuploader](https://github.com/arduino/arduino-fwuploader). The purpose of this plugin is to abstract all the
business logic needed to update firmware and certificates for the nina family boards:
- [MKR WiFi 1010](https://docs.arduino.cc/hardware/mkr-wifi-1010)
- [Nano 33 IoT](https://docs.arduino.cc/hardware/nano-33-iot)
- [UNO WiFi Rev2](https://docs.arduino.cc/hardware/uno-wifi-rev2)
- [Nano RP2040 Connect](https://docs.arduino.cc/hardware/nano-rp2040-connect)

## How to contribute

Contributions are welcome!

:sparkles: Thanks to all our [contributors](https://github.com/arduino/nina-fwuploader-plugin/graphs/contributors)! :sparkles:

### Requirements

1. [Go](https://go.dev/) version 1.20 or later
1. [Task](https://taskfile.dev/) to help you run the most common tasks from the command line
1. One of the nina family supported boards.

## Development

When running the plugin inside the fwuploader, the required tools are downloaded by the fwuploader. If you run only the plugin, you must provide them by hand.
Therefore be sure to place the `rp2040tools`, `avrdude`, and `bossac` binaries in the correct folders like the following:

```bash
.
├── bossac
│   └── 1.7.0-arduino3
│       └── bossac
├── avrdude
│   └── 6.3.0-arduino17
│       └── bin
│           └── avrdude
├── rp2040tools
│   └── 1.0.6
│       └── rp2040load
└── nina-fwuploader-plugin_linux_amd64
    └── bin
        └── nina-fwuploader-plugin
```

**Commands**

- `nina-fwuploader-plugin cert flash -p /dev/ttyACM0 ~/Documents/certificate.pem`
- `nina-fwuploader-plugin firmware get-version -p /dev/ttyACM0`
- `nina-fwuploader-plugin firmware flash -p /dev/ttyACM0 ~/Documents/fw0.2.0.bin`

## Security

If you think you found a vulnerability or other security-related bug in the nina-fwuploader-plugin, please read our [security
policy] and report the bug to our Security Team 🛡️ Thank you!

e-mail contact: security@arduino.cc

## License

nina-fwuploader-plugin is licensed under the [AGPL 3.0](LICENSE.txt) license.

You can be released from the requirements of the above license by purchasing a commercial license. Buying such a license
is mandatory if you want to modify or otherwise use the software for commercial activities involving the Arduino
software without disclosing the source code of your own applications. To purchase a commercial license, send an email to
license@arduino.cc
