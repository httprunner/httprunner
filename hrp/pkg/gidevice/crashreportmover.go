package gidevice

import (
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"

	"howett.net/plist"

	"github.com/httprunner/httprunner/v4/hrp/pkg/gidevice/pkg/libimobiledevice"
)

var _ CrashReportMover = (*crashReportMover)(nil)

func newCrashReportMover(client *libimobiledevice.CrashReportMoverClient) *crashReportMover {
	return &crashReportMover{
		client: client,
	}
}

type crashReportMover struct {
	client *libimobiledevice.CrashReportMoverClient
	afc    Afc
}

func (c *crashReportMover) readPing() (err error) {
	var data []byte
	if data, err = c.client.InnerConn().Read(4); err != nil {
		return err
	}
	if string(data) != "ping" {
		return fmt.Errorf("crashReportMover ping: %v", data)
	}

	return
}

func (c *crashReportMover) Move(hostDir string, opts ...CrashReportMoverOption) (err error) {
	opt := defaultCrashReportMoverOption()
	for _, fn := range opts {
		fn(opt)
	}

	toExtract := make([]string, 0, 64)

	fn := func(cwd string, info *AfcFileInfo) {
		if info.IsDir() {
			return
		}
		if cwd == "." {
			cwd = ""
		}

		devFilename := path.Join(cwd, info.Name())
		hostElem := strings.Split(devFilename, "/")
		hostFilename := filepath.Join(hostDir, filepath.Join(hostElem...))
		hostFilename = strings.TrimSuffix(hostFilename, ".synced")

		if opt.extract && strings.HasSuffix(hostFilename, ".plist") {
			toExtract = append(toExtract, hostFilename)
		}

		var afcFile *AfcFile
		if afcFile, err = c.afc.Open(devFilename, AfcFileModeRdOnly); err != nil {
			debugLog(fmt.Sprintf("crashReportMover open %s: %s", devFilename, err))
			return
		}
		defer func() {
			if err = afcFile.Close(); err != nil {
				debugLog(fmt.Sprintf("crashReportMover device file close: %s", err))
			}
		}()

		if err = os.MkdirAll(filepath.Dir(hostFilename), 0o755); err != nil {
			debugLog(fmt.Sprintf("crashReportMover mkdir %s: %s", filepath.Dir(hostFilename), err))
			return
		}
		var hostFile *os.File
		if hostFile, err = os.Create(hostFilename); err != nil {
			debugLog(fmt.Sprintf("crashReportMover create %s: %s", hostFilename, err))
			return
		}
		defer func() {
			if err = hostFile.Close(); err != nil {
				debugLog(fmt.Sprintf("crashReportMover host file close: %s", err))
			}
		}()

		if _, err = io.Copy(hostFile, afcFile); err != nil {
			debugLog(fmt.Sprintf("crashReportMover copy %s", err))
			return
		}

		opt.whenDone(devFilename)

		if opt.keep {
			return
		}

		if err = c.afc.Remove(devFilename); err != nil {
			debugLog(fmt.Sprintf("crashReportMover remove %s: %s", devFilename, err))
			return
		}
	}
	if err = c.walkDir(".", fn); err != nil {
		return err
	}

	if !opt.extract {
		return nil
	}

	for _, name := range toExtract {
		data, err := os.ReadFile(name)
		if err != nil {
			debugLog(fmt.Sprintf("crashReportMover extract read %s: %s", name, err))
			continue
		}
		m := make(map[string]interface{})
		if _, err = plist.Unmarshal(data, &m); err != nil {
			debugLog(fmt.Sprintf("crashReportMover extract plist %s: %s", name, err))
			continue
		}

		desc, ok := m["description"]
		if !ok {
			continue
		}
		hostExtCrash := strings.TrimSuffix(name, ".plist") + ".crash"
		if err = os.WriteFile(hostExtCrash, []byte(fmt.Sprintf("%v", desc)), 0o755); err != nil {
			debugLog(fmt.Sprintf("crashReportMover extract save %s: %s", name, err))
			continue
		}
	}

	return
}

func (c *crashReportMover) walkDir(dirname string, fn func(path string, info *AfcFileInfo)) (err error) {
	var names []string
	if names, err = c.afc.ReadDir(dirname); err != nil {
		return err
	}

	cwd := dirname

	for _, n := range names {
		if n == "." || n == ".." {
			continue
		}

		var info *AfcFileInfo
		if info, err = c.afc.Stat(path.Join(cwd, n)); err != nil {
			return err
		}
		if info.IsDir() {
			if err = c.walkDir(path.Join(cwd, info.name), fn); err != nil {
				return err
			}
		}

		fn(cwd, info)
	}

	return
}
