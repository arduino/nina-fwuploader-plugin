package main

import (
	helper "github.com/arduino/fwuploader-plugin-helper"
	paths "github.com/arduino/go-paths-helper"
	semver "go.bug.st/relaxed-semver"
)

func main() {
	helper.RunPlugin(&ninaPlugin{})
}

type ninaPlugin struct {
}

var _ helper.Plugin = (*ninaPlugin)(nil)

// GetFirmwareVersion implements helper.Plugin.
func (*ninaPlugin) GetFirmwareVersion(portAddress string, fqbn string, feedback *helper.PluginFeedback) (*semver.RelaxedVersion, error) {
	panic("unimplemented")
}

// GetPluginInfo implements helper.Plugin.
func (*ninaPlugin) GetPluginInfo() *helper.PluginInfo {
	panic("unimplemented")
}

// UploadCertificate implements helper.Plugin.
func (*ninaPlugin) UploadCertificate(portAddress string, fqbn string, certificatePath *paths.Path, feedback *helper.PluginFeedback) error {
	panic("unimplemented")
}

// UploadFirmware implements helper.Plugin.
func (*ninaPlugin) UploadFirmware(portAddress string, fqbn string, firmwarePath *paths.Path, feedback *helper.PluginFeedback) error {
	panic("unimplemented")
}
