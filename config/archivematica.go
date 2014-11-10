// +build archivematica

package config

func init() {
	siegfried.home = "/usr/share/siegfried"
	siegfried.signature = "archivematica.gob"
	siegfried.signatureVersion = 1
	identifier.name = "archivematica"
}
