package ghdc

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"
)

type uitestTransport struct {
	connectionPool *ConnectionPool
	readTimeout    time.Duration
}

func newUitestTransport(host string, port string, readTimeout ...time.Duration) (uTp uitestTransport, err error) {
	if len(readTimeout) == 0 {
		readTimeout = []time.Duration{DefaultReadTimeout}
	}
	uTp.readTimeout = readTimeout[0]
	pool, err := newConnectionPool(host, port, 2)
	if err != nil {
		err = fmt.Errorf("[uitest] failed to init connection pool \n%v", err)
		return
	}
	uTp.connectionPool = pool
	return
}

func newHypiumRequest(params interface{}, method string) UitestRequest {
	return UitestRequest{
		Module:    "com.ohos.devicetest.hypiumApiHelper",
		Method:    method,
		Params:    params,
		RequestId: MD5(fmt.Sprintf("%d", time.Now().UnixNano())),
	}
}

func (uTp *uitestTransport) SendReq(req UitestRequest) (res UitestResponse, err error) {
	requestBytes, err := json.Marshal(req)
	if err != nil {
		err = fmt.Errorf("[uitest] failed to marshal %v request %v", req, err)
		return
	}
	requestBytes = append(requestBytes, '\n')
	conn, err := uTp.connectionPool.getConnection()
	if err != nil {
		err = fmt.Errorf("[uitest] failed to get connection \n%v", err)
		return
	}
	defer func() {
		uTp.connectionPool.releaseConnection(conn)
	}()
	_ = conn.SetReadDeadline(time.Now().Add(uTp.readTimeout))
	debugLog(fmt.Sprintf("[uitest] send Request %v", req))
	err = _send(conn, requestBytes)
	if err != nil {
		err = fmt.Errorf("[uitest] failed to get send %v request %v", req, err)
		return
	}
	raw, err := _read(conn)
	if err != nil {
		err = fmt.Errorf("[uitest] failed to read %v response %v", req, err)
		return
	}
	res = UitestResponse{}
	err = json.Unmarshal(raw, &res)
	return
}

func (uTp *uitestTransport) Close() {
	if uTp.connectionPool != nil {
		uTp.connectionPool.close()
	}
	uTp.connectionPool = nil
}

func MD5(str string) string {
	hash := md5.New()
	hash.Write([]byte(str))
	return hex.EncodeToString(hash.Sum(nil))
}
