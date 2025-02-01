package common

import "runtime"

var (
	version   = "5.6.1.1"
	gitcommit = "unknown"
	buildtime = "2025-02-03T12:29:0.920912-09:00"
)

// VersionInfo holds information about the running Blockbook instance
type VersionInfo struct {
	Version   string `json:"version"`
	GitCommit string `json:"gitcommit"`
	BuildTime string `json:"buildtime"`
	GoVersion string `json:"goversion"`
	OSArch    string `json:"os/arch"`
}

// GetVersionInfo returns VersionInfo of the running Blockbook instance
func GetVersionInfo() VersionInfo {
	return VersionInfo{
		Version:   version,
		GitCommit: gitcommit,
		BuildTime: buildtime,
		GoVersion: runtime.Version(),
		OSArch:    runtime.GOOS + "/" + runtime.GOARCH,
	}
}
