package pronom

import (
	"path/filepath"
	"testing"

	"github.com/richardlehane/siegfried/pkg/config"
)

var p *pronom

func TestNew(t *testing.T) {
	config.SetHome(filepath.Join("..", "..", "cmd", "roy", "data"))
	_, err := NewPronom()
	if err != nil {
		t.Error(err)
	}
}
