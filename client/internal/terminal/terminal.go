package terminal

import (
	_ "embed"
	"os"
	"os/exec"

	"github.com/creack/pty"
	log "github.com/sirupsen/logrus"
)

type terminal struct {
	width, height uint16

	ptmx *os.File
}

// NewTerminal implement close of terminal
func newTerminal(w uint16, h uint16) (*terminal, error) {
	t := &terminal{
		width:  w,
		height: h,
	}
	err := t.open()
	if err != nil {
		return nil, err
	}

	return t, nil
}

func (t *terminal) open() error {
	var err error

	for _, cmd := range shells {
		c := exec.Command(cmd)
		c.Env = os.Environ()
		c.Env = append(c.Env, envs...)
		t.ptmx, err = pty.Start(c)
		if err == nil {
			break
		}
	}

	if err != nil {
		return err
	}

	t.updateSize()
	return err
}

func (t *terminal) read(buf []byte) (int, error) {
	return t.ptmx.Read(buf)
}

func (t *terminal) write(data []byte, width, height uint16) {
	if t.width != width || t.height != height {
		t.width = width
		t.height = height
	}
	t.updateSize()
	_, err := t.ptmx.Write(data)
	if err != nil {
		log.Errorf("failed to write out: %s", err)
	}
}

func (t *terminal) updateSize() {
	size := &pty.Winsize{
		Cols: t.width,
		Rows: t.height,
	}
	err := pty.Setsize(t.ptmx, size)
	if err != nil {
		log.Debugf("failed to resize pty: %s", err)
	}
}
