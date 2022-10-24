package gidevice

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	"github.com/httprunner/httprunner/v4/hrp/pkg/gidevice/pkg/libimobiledevice"
)

var _ SyslogRelay = (*syslogRelay)(nil)

func newSyslogRelay(client *libimobiledevice.SyslogRelayClient) *syslogRelay {
	r := &syslogRelay{
		client:    client,
		stop:      make(chan bool),
		isReading: false,
	}
	r.reader = bufio.NewReader(r.client.InnerConn().RawConn())
	return r
}

type syslogRelay struct {
	client *libimobiledevice.SyslogRelayClient

	reader    *bufio.Reader
	stop      chan bool
	isReading bool
}

func (r *syslogRelay) Lines() <-chan string {
	out := make(chan string)
	r.isReading = true

	go func() {
		defer func() {
			close(out)
			r.isReading = false
		}()
		for {
			select {
			case <-r.stop:
				return
			default:
				bs, err := r.readLine()
				if err != nil {
					if strings.Contains(err.Error(), io.EOF.Error()) {
						return
					}
					debugLog(fmt.Sprintf("syslog: %s", err))
				}
				if len(bs) > 1 && bs[0] == 0 {
					bs = bs[1:]
				}
				out <- string(bs)
			}
		}
	}()
	return out
}

func (r *syslogRelay) Stop() {
	if r.isReading {
		r.stop <- true
	}
}

func (r *syslogRelay) readLine() ([]byte, error) {
	var line []byte
	for {
		l, more, err := r.reader.ReadLine()
		if err != nil {
			return nil, err
		}
		if line == nil && !more {
			return l, nil
		}
		line = append(line, l...)
		if !more {
			break
		}
	}
	return line, nil
}
