package main

import (
	"embed"
	"os"

	helper "github.com/arduino/fwuploader-plugin-helper"
	paths "github.com/arduino/go-paths-helper"
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

	//go:embed sketches/commands/build/arduino.mbed_nano.nanorp2040connect/commands.ino.bin
	nanorp2040connectCommandSketchBinary embed.FS

	//go:embed sketches/commands/build/arduino.megaavr.uno2018/commands.ino.bin
	uno2018CommandSketchBinary embed.FS

	//go:embed sketches/commands/build/arduino.samd.mkrvidor4000/commands.ino.bin
	mkrvidor4000CommandSketchBinary embed.FS

	//go:embed sketches/commands/build/arduino.samd.mkrwifi1010/commands.ino.bin
	mkrwifi1010CommandSketchBinary embed.FS

	//go:embed sketches/commands/build/arduino.samd.nano_33_iot/commands.ino.bin
	nano33iotCommandSketchBinary embed.FS
)

func main() {
	espflashPath, err := helper.FindToolPath("espflash", semver.MustParse("2.0.1"))
	if err != nil {
		slog.Error("Couldn't find espflash@2.0.1 binary")
		os.Exit(1)
	}
	bossacPath, err := helper.FindToolPath("bossac", semver.MustParse("1.9.1-arduino5"))
	if err != nil {
		slog.Error("Couldn't find bossac@1.9.1-arduino5 binary")
		os.Exit(1)
	}

	helper.RunPlugin(&ninaPlugin{
		espflashBin: espflashPath,
		bossacBin:   bossacPath,
	})
}

type ninaPlugin struct {
	espflashBin *paths.Path
	bossacBin   *paths.Path
}

var _ helper.Plugin = (*ninaPlugin)(nil)

// GetFirmwareVersion implements helper.Plugin.
func (*ninaPlugin) GetFirmwareVersion(portAddress string, fqbn string, feedback *helper.PluginFeedback) (*semver.RelaxedVersion, error) {
	panic("unimplemented")
}

// GetPluginInfo implements helper.Plugin.
func (*ninaPlugin) GetPluginInfo() *helper.PluginInfo {
	return &helper.PluginInfo{
		Name:    pluginName,
		Version: semver.MustParse(versionString),
	}
}

// UploadCertificate implements helper.Plugin.
func (*ninaPlugin) UploadCertificate(portAddress string, fqbn string, certificatePath *paths.Path, feedback *helper.PluginFeedback) error {
	panic("unimplemented")
}

// UploadFirmware implements helper.Plugin.
func (*ninaPlugin) UploadFirmware(portAddress string, fqbn string, firmwarePath *paths.Path, feedback *helper.PluginFeedback) error {
	panic("unimplemented")
}
