package main

import (
	"bufio"
	"embed"
	"fmt"
	"os"
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
	espflashPath, err := helper.FindToolPath("espflash", semver.MustParse("2.0.0"))
	if err != nil {
		slog.Error("Couldn't find espflash@2.0.0 binary")
		os.Exit(1)
	}
	bossacPath, err := helper.FindToolPath("bossac", semver.MustParse("1.7.0-arduino3"))
	if err != nil {
		slog.Error("Couldn't find bossac@1.7.0-arduino3 binary")
		os.Exit(1)
	}

	helper.RunPlugin(&ninaPlugin{
		espflashBin: espflashPath.Join("espflash"),
		bossacBin:   bossacPath.Join("bossac"),
	})
}

type ninaPlugin struct {
	espflashBin *paths.Path
	bossacBin   *paths.Path
}

var _ helper.Plugin = (*ninaPlugin)(nil)

// GetFirmwareVersion implements helper.Plugin.
func (p *ninaPlugin) GetFirmwareVersion(portAddress string, fqbn string, feedback *helper.PluginFeedback) (*semver.RelaxedVersion, error) {
	if err := p.uploadCommandsSketch(portAddress, fqbn, feedback); err != nil {
		return nil, err
	}
	// TODO should we check for possible port change, that might occur after the sketch upload?
	port, err := serial.Open(portAddress)
	if err != nil {
		return nil, err
	}
	defer port.Close()

	if _, err := port.Write([]byte("v")); err != nil {
		return nil, fmt.Errorf("write to serial port: %v", err)
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

	if err := p.uploadCommandsSketch(portAddress, fqbn, feedback); err != nil {
		return err
	}

	if err := p.reboot(&portAddress, feedback); err != nil {
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

	cmd, err := executils.NewProcess([]string{}, p.espflashBin.String(), "write-bin", "-p", portAddress, "-b", "1000000", "0x10000", certData.String())
	if err != nil {
		return err
	}
	cmd.RedirectStderrTo(feedback.Err())
	cmd.RedirectStdoutTo(feedback.Out())
	if err := cmd.Run(); err != nil {
		return err
	}

	fmt.Fprintf(feedback.Out(), "\nUpload completed! To complete the update process please disconnect and then reconnect the board.\n")
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

	if err := p.uploadCommandsSketch(portAddress, fqbn, feedback); err != nil {
		return err
	}

	if err := p.reboot(&portAddress, feedback); err != nil {
		return err
	}

	cmd, err := executils.NewProcess(nil, p.espflashBin.String(), "write-bin", "-p", portAddress, "-b", "1000000", "0x0", firmwarePath.String())
	if err != nil {
		return err
	}
	cmd.RedirectStderrTo(feedback.Err())
	cmd.RedirectStdoutTo(feedback.Out())
	if err := cmd.Run(); err != nil {
		return err
	}

	fmt.Fprintf(feedback.Out(), "\nUpload completed! To complete the update process please disconnect and then reconnect the board.\n")

	return nil
}

func (p *ninaPlugin) uploadCommandsSketch(portAddress, fqbn string, feedback *helper.PluginFeedback) error {
	slog.Info("upload command sketch")

	readCommandsSketch := func(fqbn string) ([]byte, error) {
		switch fqbn {
		case "arduino:mbed_nano:nanorp2040connect":
			return commandSketchDir.ReadFile("sketches/commands/build/arduino.mbed_nano.nanorp2040connect/commands.ino.bin")
		case "arduino:megaavr:uno2018":
			return commandSketchDir.ReadFile("sketches/commands/build/arduino.megaavr.uno2018/commands.ino.bin")
		case "arduino:samd:mkrvidor4000":
			return commandSketchDir.ReadFile("sketches/commands/build/arduino.samd.mkrvidor4000/commands.ino.bin")
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
	if err := serialutils.TouchSerialPortAt1200bps(portAddress); err != nil {
		return err
	}

	slog.Info("uploading command sketch with bossac")
	cmd, err := executils.NewProcess(nil, p.bossacBin.String(), "-i", "-d", "--port="+portAddress, "-U", "true", "-i", "-e", "-w", "-v", rebootFile.String(), "-R")
	if err != nil {
		return err
	}
	cmd.RedirectStderrTo(feedback.Err())
	cmd.RedirectStdoutTo(feedback.Out())
	if err := cmd.Run(); err != nil {
		return err
	}

	time.Sleep(3 * time.Second)
	return nil
}

func (p *ninaPlugin) reboot(portAddress *string, feedback *helper.PluginFeedback) error {
	// Will be used later to check if the OS changed the serial port.
	allSerialPorts, err := serial.AllPorts()
	if err != nil {
		return err
	}

	fmt.Fprintf(feedback.Out(), "\nWaiting to flash the binary...\n")

	port, err := serial.Open(*portAddress)
	if err != nil {
		return err
	}

	if err := serial.SendCommandAndClose(port, serial.Command("r")); err != nil {
		return err
	}

	slog.Info("check if serial port has changed")
	// When a board is successfully rebooted in esp32 mode, it might change the serial port.
	// Every 250ms we're watching for new ports, if a new one is found we return that otherwise
	// we'll wait the 10 seconds timeout expiration.
	newPort, changed, err := allSerialPorts.NewPort()
	if err != nil {
		return err
	}
	if changed {
		*portAddress = newPort
	}
	return nil
}
