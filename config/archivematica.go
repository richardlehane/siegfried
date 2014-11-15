// +build archivematica

package config

func init() {
	siegfried.home = "/usr/share/siegfried"
	siegfried.signature = "archivematica.gob"
	siegfried.signatureVersion = 1
	identifier.name = "archivematica"
	pronom.extend = []string{"archivematica-fmt/2", "archivematica-fmt/3", "archivematica-fmt/4", "archivematica-fmt/5"}
}
