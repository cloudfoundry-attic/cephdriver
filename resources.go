package cephdriver

type MountConfig struct {
	Keyring    string `json:"keyring"`
	IP         string `json:"ip"`
	MountPoint string `json:"mountpoint"`
}
