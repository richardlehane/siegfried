// +build archivematica

package config

func init() {
	Siegfried.Home = "/usr/share/siegfried"
	Siegfried.Signature = "archivematica.gob"
	Siegfried.SignatureVersion = 1
	Identifier.Name = "archivematica"
}
