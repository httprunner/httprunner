package uixt

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/httprunner/httprunner/v4/hrp/internal/json"
)

type ShootsAndroidDriver struct {
	socket  net.Conn
	seq     int
	timeout time.Duration
	adbDriver
}

const ShootsSocketName = "com.bytest.device"

// newShootsAndroidDriver
// 创建shoots Driver address为forward后的端口格式127.0.0.1:${port}
func newShootsAndroidDriver(address string, readTimeout ...time.Duration) (*ShootsAndroidDriver, error) {
	timeout := 10 * time.Second
	if len(readTimeout) > 0 {
		timeout = readTimeout[0]
	}

	conn, err := net.Dial("tcp", address)
	if err != nil {
		log.Err(err).Msg(fmt.Sprintf("failed to connect %s", address))
		return nil, err
	}

	return &ShootsAndroidDriver{
		socket:  conn,
		timeout: timeout,
	}, nil
}

func (sad *ShootsAndroidDriver) NewSession(capabilities Capabilities) (SessionInfo, error) {
	return SessionInfo{}, errDriverNotImplemented
}

func (sad *ShootsAndroidDriver) sendCommand(packageName string, cmdType string, params map[string]interface{}, readTimeout ...time.Duration) (interface{}, error) {
	sad.seq++
	packet := map[string]interface{}{
		"Seq": sad.seq,
		"Cmd": cmdType,
		"v":   "",
	}
	for key, value := range params {
		if key == "Cmd" || key == "Seq" {
			return "", errors.New("params cannot be Cmd or Seq")
		}
		packet[key] = value
	}
	data, err := json.Marshal(packet)
	if err != nil {
		return nil, err
	}

	res, err := sad.adbClient.RunShootsCommand(append(data, '\n'), packageName)
	if err != nil {
		return nil, err
	}
	var resultMap map[string]interface{}
	if err := json.Unmarshal([]byte(res), &resultMap); err != nil {
		return nil, err
	}
	if resultMap["Error"] != nil {
		return nil, fmt.Errorf("failed to call shoots command: %s", resultMap["Error"].(string))
	}

	return resultMap["Result"], nil
}

func (sad *ShootsAndroidDriver) send(data []byte, readTimeout ...time.Duration) (map[string]interface{}, error) {
	timeout := sad.timeout
	if len(readTimeout) > 0 {
		timeout = readTimeout[0]
	}
	_ = sad.socket.SetReadDeadline(time.Now().Add(timeout))

	err := _send(sad.socket, append(data, '\n'))
	if err != nil {
		sad.close()
		return nil, err
	}
	raw, err := _readAll(sad.socket)
	if err != nil {
		return nil, err
	}
	var result map[string]interface{}
	if err := json.Unmarshal(raw, &result); err != nil {
		log.Printf("error when parse json response: %s\n", raw)
		return nil, err
	}
	return result, nil
}

func _send(writer io.Writer, msg []byte) (err error) {
	for totalSent := 0; totalSent < len(msg); {
		var sent int
		if sent, err = writer.Write(msg[totalSent:]); err != nil {
			return err
		}
		if sent == 0 {
			return errors.New("socket connection broken")
		}
		totalSent += sent
	}
	return
}

func _readN(reader io.Reader, size int) (raw []byte, err error) {
	raw = make([]byte, 0, size)
	for len(raw) < size {
		buf := make([]byte, size-len(raw))
		var n int
		if n, err = io.ReadFull(reader, buf); err != nil {
			return nil, err
		}
		if n == 0 {
			return nil, errors.New("socket connection broken")
		}
		raw = append(raw, buf...)
	}
	return
}

func _readAll(reader io.Reader) (raw []byte, err error) {
	buffer := new(bytes.Buffer)
	for true {
		lengthBuf := make([]byte, 4)
		_, err := io.ReadFull(reader, lengthBuf)
		if err != nil {
			if err == io.EOF {
				return buffer.Bytes(), nil
			} else if errors.Is(err, io.ErrUnexpectedEOF) {
				err = fmt.Errorf("reached unexpected EOF, read partial data: %s %v", string(buffer.Bytes()), err)
				return nil, err
			} else {
				return nil, err
			}
		}
		length := binary.BigEndian.Uint32(lengthBuf)

		data, err := _readN(reader, int(length)-4)
		if err != nil {
			return nil, err
		}
		buffer.Write(data)

	}
	return buffer.Bytes(), nil
}

func (sad *ShootsAndroidDriver) DeleteSession() error {
	return sad.close()
}

func (sad *ShootsAndroidDriver) close() error {
	if sad.socket != nil {
		return sad.socket.Close()
	}
	return nil
}

func (sad *ShootsAndroidDriver) Status() (DeviceStatus, error) {
	app, err := sad.GetForegroundApp()
	res, err := sad.sendCommand(app.PackageName, "Hello", nil)
	if err != nil {
		return DeviceStatus{}, err
	}
	log.Info().Msg(fmt.Sprintf("pint shoots result :%v", res))
	return DeviceStatus{}, nil
}

func (sad *ShootsAndroidDriver) Source(srcOpt ...SourceOption) (source string, err error) {
	app, err := sad.GetForegroundApp()
	params := map[string]interface{}{
		"ClassName": "com.bytedance.byteinsight.MockOperator",
		"Method":    "getLayout",
		"RetType":   "",
		"Args":      []string{},
	}
	res, err := sad.sendCommand(app.PackageName, "CallStaticMethod", params)
	if err != nil {
		return "", err
	}
	return res.(string), nil
}

func (sad *ShootsAndroidDriver) LoginNoneUI(packageName, phoneNumber string, captcha string) error {
	_, err := sad.adbClient.RunShellCommand("am", "broadcast", "-a", fmt.Sprintf("%s.util.crony.action_login", packageName), "-e", "phone", phoneNumber, "-e", "code", captcha)
	time.Sleep(5 * time.Second)
	login, err := sad.isLogin(packageName)
	if err != nil || !login {
		log.Err(err).Msg("failed to login")
		return fmt.Errorf("failed to login")
	}
	return err
}

func (sad *ShootsAndroidDriver) isLogin(packageName string) (login bool, err error) {
	params := map[string]interface{}{
		"ClassName":   "com.ss.android.ugc.aweme.account.AccountProxyService",
		"Method":      "userService",
		"RetType":     "",
		"Args":        []string{},
		"CacheObject": true,
	}
	id, err := sad.sendCommand(packageName, "CallStaticMethod", params)
	if err != nil {
		return false, err
	}

	params = map[string]interface{}{
		"Method":   "isLogin",
		"RetType":  "",
		"Args":     []string{},
		"ObjectId": int(id.(float64)),
	}
	loginObj, err := sad.sendCommand(packageName, "CallMethod", params)
	if err != nil {
		return false, err
	}
	return loginObj.(bool), nil
}
