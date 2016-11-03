// +build archivematica

package config

func init() {
	siegfried.home = "/usr/share/siegfried"
	siegfried.signature = "archivematica.sig"
	identifier.name = "archivematica"
	identifier.extend = []string{"archivematica-fmt2.xml", "archivematica-fmt3.xml", "archivematica-fmt4.xml", "archivematica-fmt5.xml"}
}
