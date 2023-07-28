package main

import (
	"bufio"
	"embed"
	"fmt"
	"time"

	"github.com/arduino/arduino-cli/executils"
	helper "github.com/arduino/fwuploader-plugin-helper"
	paths "github.com/arduino/go-paths-helper"
	"github.com/arduino/uno-r4-wifi-fwuploader-plugin/serial"
	serialutils "github.com/arduino/uno-r4-wifi-fwuploader-plugin/serial/utils"
	semver "go.bug.st/relaxed-semver"
	"golang.org/x/exp/slog"
)

const (
	pluginName = "nina-fwuploader"
)

var (
	versionString = "0.0.0-git"
	commit        = ""
	date          = ""

	//go:embed sketches/commands/build
	commandSketchDir embed.FS
)

func main() {
	helper.RunPlugin(&ninaPlugin{})
}

type ninaPlugin struct{}

var _ helper.Plugin = (*ninaPlugin)(nil)

// GetFirmwareVersion implements helper.Plugin.
func (p *ninaPlugin) GetFirmwareVersion(portAddress string, fqbn string, feedback *helper.PluginFeedback) (*semver.RelaxedVersion, error) {
	if err := p.uploadCommandsSketch(&portAddress, fqbn, feedback); err != nil {
		return nil, err
	}

	port, err := serialOpen(portAddress)
	if err != nil {
		return nil, err
	}
	defer port.Close()

	// be sure that the serial port is ready
	time.Sleep(2 * time.Second)

	if err := getVersion(port); err != nil {
		return nil, err
	}

	var version string
	scanner := bufio.NewScanner(port)
	for scanner.Scan() {
		version = scanner.Text()
		break
	}

	return semver.ParseRelaxed(version), nil
}

// GetPluginInfo implements helper.Plugin.
func (p *ninaPlugin) GetPluginInfo() *helper.PluginInfo {
	return &helper.PluginInfo{
		Name:    pluginName,
		Version: semver.MustParse(versionString),
	}
}

// UploadCertificate implements helper.Plugin.
func (p *ninaPlugin) UploadCertificate(portAddress string, fqbn string, certificatePath *paths.Path, feedback *helper.PluginFeedback) error {
	if portAddress == "" {
		return fmt.Errorf("invalid port address")
	}
	if certificatePath == nil || certificatePath.IsDir() || !certificatePath.Exist() {
		return fmt.Errorf("invalid certificate path")
	}
	fmt.Fprintf(feedback.Out(), "Uploading certificates to %s...\n", portAddress)

	if err := p.uploadCommandsSketch(&portAddress, fqbn, feedback); err != nil {
		return err
	}

	certificatesData, err := certificatePath.ReadFile()
	if err != nil {
		return err
	}

	// The certificate must be multiple of 4, otherwise `espflash` won't work!
	// (https://github.com/esp-rs/espflash/issues/440)
	for len(certificatesData)&3 != 0 {
		certificatesData = append(certificatesData, 0)
	}

	certificatesDataLimit := 0x20000
	if len(certificatesData) > certificatesDataLimit {
		return fmt.Errorf("certificates data %d exceeds limit of %d bytes", len(certificatesData), certificatesDataLimit)
	}

	certData, err := paths.WriteToTempFile(certificatesData, paths.TempDir(), "nina-fwuploader-plugin-cert")
	if err != nil {
		return err
	}
	defer certData.Remove()

	flasher, err := newFlasher(portAddress)
	if err != nil {
		return err
	}

	if err := flasher.flashChunk(0x10000, certificatesData); err != nil {
		return err
	}

	fmt.Fprintf(feedback.Out(), "\nUpload completed!")
	return nil
}

// UploadFirmware implements helper.Plugin.
func (p *ninaPlugin) UploadFirmware(portAddress string, fqbn string, firmwarePath *paths.Path, feedback *helper.PluginFeedback) error {
	if portAddress == "" {
		return fmt.Errorf("invalid port address")
	}
	if firmwarePath == nil || firmwarePath.IsDir() || !firmwarePath.Exist() {
		return fmt.Errorf("invalid firmware path")
	}

	if err := p.uploadCommandsSketch(&portAddress, fqbn, feedback); err != nil {
		return err
	}

	flasher, err := newFlasher(portAddress)
	if err != nil {
		return err
	}

	fwData, err := firmwarePath.ReadFile()
	if err != nil {
		return err
	}
	if err := flasher.flashChunk(0x0000, fwData); err != nil {
		return err
	}

	if err := flasher.md5sum(fwData); err != nil {
		return err
	}

	fmt.Fprintf(feedback.Out(), "\nUpload completed!")

	return nil
}

func (p *ninaPlugin) uploadCommandsSketch(portAddress *string, fqbn string, feedback *helper.PluginFeedback) error {
	slog.Info("upload command sketch")

	readCommandsSketch := func(fqbn string) ([]byte, error) {
		switch fqbn {
		case "arduino:mbed_nano:nanorp2040connect":
			return commandSketchDir.ReadFile("sketches/commands/build/arduino.mbed_nano.nanorp2040connect/commands.ino.elf")
		case "arduino:megaavr:uno2018":
			return commandSketchDir.ReadFile("sketches/commands/build/arduino.megaavr.uno2018/commands.ino.hex")
		case "arduino:samd:mkrwifi1010":
			return commandSketchDir.ReadFile("sketches/commands/build/arduino.samd.mkrwifi1010/commands.ino.bin")
		case "arduino:samd:nano_33_iot":
			return commandSketchDir.ReadFile("sketches/commands/build/arduino.samd.nano_33_iot/commands.ino.bin")
		default:
			return nil, fmt.Errorf("the board (fqbn: %v) is not supported", fqbn)
		}
	}

	sketchData, err := readCommandsSketch(fqbn)
	if err != nil {
		return err
	}

	rebootFile, err := paths.WriteToTempFile(sketchData, paths.TempDir(), "nina-fwuploader-plugin")
	if err != nil {
		return err
	}
	defer rebootFile.Remove()

	slog.Info("sending serial reset")

	// Will be used after the 1200 touch to check if the OS changed the serial port.
	allSerialPorts, err := serial.AllPorts()
	if err != nil {
		return err
	}

	if err := serialutils.TouchSerialPortAt1200bps(*portAddress); err != nil {
		return err
	}

	newPort, changed, err := allSerialPorts.NewPort()
	if err != nil {
		return err
	}
	if changed {
		*portAddress = newPort
	}

	getSketchUploader := func(fqbn string) (*executils.Process, error) {
		switch fqbn {
		case "arduino:mbed_nano:nanorp2040connect":
			rp2040loadPath, err := helper.FindToolPath("rp2040tools", semver.MustParse("1.0.6"))
			if err != nil {
				return nil, fmt.Errorf("couldn't find rp2040tools@1.0.6 binary")
			}

			return executils.NewProcess(nil, rp2040loadPath.Join("rp2040load").String(), "-v", "-D", rebootFile.String())
		case "arduino:megaavr:uno2018":
			avrdudePath, err := helper.FindToolPath("avrdude", semver.MustParse("6.3.0-arduino17"))
			if err != nil {
				return nil, fmt.Errorf("couldn't find avrdude@6.3.0-arduino17 binary")
			}

			return executils.NewProcess(nil, avrdudePath.Join("bin", "avrdude").String(), "-C"+avrdudePath.Join("etc", "avrdude.conf").String(), "-v", "-patmega4809", "-cxplainedmini_updi", "-Pusb", "-b115200", "-e", "-D", fmt.Sprintf("-Uflash:w:%v:i", rebootFile.String()), "-Ufuse2:w:0x01:m", "-Ufuse5:w:0xC9:m", "-Ufuse8:w:0x02:m")
		case "arduino:samd:mkrwifi1010", "arduino:samd:nano_33_iot":
			bossacPath, err := helper.FindToolPath("bossac", semver.MustParse("1.7.0-arduino3"))
			if err != nil {
				return nil, fmt.Errorf("couldn't find bossac@1.7.0-arduino3 binary")
			}

			return executils.NewProcess(nil, bossacPath.Join("bossac").String(), "-i", "-d", "--port="+*portAddress, "-U", "true", "-i", "-e", "-w", "-v", rebootFile.String(), "-R")
		}
		return nil, fmt.Errorf("sketch uploader: the board (fqbn: %v) is not supported", fqbn)
	}

	slog.Info("uploading command sketch")
	cmd, err := getSketchUploader(fqbn)
	if err != nil {
		return err
	}
	cmd.RedirectStderrTo(feedback.Err())
	cmd.RedirectStdoutTo(feedback.Out())
	if err := cmd.Run(); err != nil {
		return err
	}

	slog.Info("upload command sketch completed")
	time.Sleep(3 * time.Second)
	return nil
}
