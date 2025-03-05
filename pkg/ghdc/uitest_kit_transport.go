package ghdc

import (
	"bytes"
	"container/list"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"net"
	"strconv"
	"sync"
	"time"
)

type uitestKitTransport struct {
	connectionPool *ConnectionPool
	socketMap      map[string]*SocketContext
	serial         string
	mu             sync.Mutex
}

type SocketContext struct {
	conn      net.Conn
	socketId  string
	writeLock sync.Mutex

	callbackMap map[string]UitestKitCallback
	queue       *responseList
}

type response struct {
	sessionId uint32
	payload   []byte
}

type responseList struct {
	list *list.List
	lock sync.Mutex
}

func (r *responseList) Clear() {
	r.lock.Lock()
	defer r.lock.Unlock()
	r.list.Init()
}

func (r *responseList) push(v interface{}) {
	defer r.lock.Unlock()
	r.lock.Lock()
	r.list.PushBack(v)
}

func (r *responseList) traverse(action func(response response)) {
	r.lock.Lock()
	defer r.lock.Unlock()
	for e := r.list.Front(); e != nil; e = e.Next() {
		action(e.Value.(response))
	}
}

func (r *responseList) remove(sessionId uint32) *response {
	r.lock.Lock()
	defer r.lock.Unlock()
	for e := r.list.Front(); e != nil; e = e.Next() {
		response := e.Value.(response)
		if response.sessionId == sessionId {
			r.list.Remove(e)
			return &response
		}
	}
	return nil
}

type UitestKitCallback interface {
	OnData([]byte)
	OnError(error)
}

type ReqTypeEnum int

const (
	DEFAULT ReqTypeEnum = iota
	SCREEN_CAPTURE
	UI_ACTION_CAPTURE
)

const (
	HEADER_BYTES = "_uitestkit_rpc_message_head_"
	TAILER_BYTES = "_uitestkit_rpc_message_tail_"
)

func newUitestKitTransport(serial string, host string, port string) (uKtp uitestKitTransport, err error) {
	pool, err := newConnectionPool(host, port, 3)
	if err != nil {
		err = fmt.Errorf("[uitest] failed to init connection pool \n%v", err)
		return
	}
	uKtp.connectionPool = pool
	uKtp.socketMap = make(map[string]*SocketContext)
	uKtp.serial = serial
	return
}

func (uKtp *uitestKitTransport) initializeSocket(reqType ReqTypeEnum) error {
	socketId := fmt.Sprintf("socket_%d_%s", reqType, uKtp.serial)
	uKtp.mu.Lock()
	defer uKtp.mu.Unlock()
	if uKtp.socketMap[socketId] != nil {
		return nil
	}
	connection, err := uKtp.connectionPool.getConnection()
	if err != nil {
		return err
	}

	uKtp.socketMap[socketId] = &SocketContext{socketId: socketId, conn: connection, callbackMap: make(map[string]UitestKitCallback), queue: newResponseList()}
	go uKtp.receive(reqType)
	return nil
}

func (uKtp *uitestKitTransport) disconnect(reqType ReqTypeEnum) {
	socketId := fmt.Sprintf("socket_%d_%s", reqType, uKtp.serial)
	uKtp.mu.Lock()
	defer uKtp.mu.Unlock()
	socketContext := uKtp.socketMap[socketId]
	if socketContext == nil {
		return
	}
	socketContext.Close()
	delete(uKtp.socketMap, socketId)
}

func (uKtp *uitestKitTransport) registerCallback(reqType ReqTypeEnum, sessionId uint32, callback UitestKitCallback) error {
	socketId := fmt.Sprintf("socket_%d_%s", reqType, uKtp.serial)
	socketContext := uKtp.socketMap[socketId]
	if socketContext == nil {
		if err := uKtp.initializeSocket(reqType); err != nil {
			return err
		}
		socketContext = uKtp.socketMap[socketId]
	}
	for {
		socketContext.writeLock.Lock()
		res := socketContext.queue.remove(sessionId)
		socketContext.writeLock.Unlock()
		if res != nil && callback != nil {
			callback.OnData(res.payload)
		}
		if res == nil {
			break
		}
	}
	socketContext.writeLock.Lock()
	socketContext.callbackMap[strconv.Itoa(int(sessionId))] = callback
	socketContext.writeLock.Unlock()
	return nil
}

func (uKtp *uitestKitTransport) receive(reqType ReqTypeEnum) {
	socketId := fmt.Sprintf("socket_%d_%s", reqType, uKtp.serial)
	defer uKtp.disconnect(reqType)
	socketContext := uKtp.socketMap[socketId]
	var receiveError error
	defer func() {
		if receiveError != nil {
			socketContext.onException(receiveError)
		}
	}()
	if socketContext == nil {
		if receiveError = uKtp.initializeSocket(reqType); receiveError != nil {
			return
		}
		socketContext = uKtp.socketMap[socketId]
	}

	for {
		headerSize := len(HEADER_BYTES)
		raw, err := _readN(socketContext.conn, headerSize+8)
		if err != nil {
			receiveError = err
			break
		}
		if len(raw) == 0 {
			continue
		}

		header := raw[:len(HEADER_BYTES)]
		if !bytes.Equal(header, []byte(HEADER_BYTES)) {
			receiveError = fmt.Errorf("verify message head failed on channel: %s", socketId)
			break
		}
		var sessionId, length uint32
		if receiveError = binary.Read(bytes.NewReader(raw[headerSize:headerSize+4]), binary.BigEndian, &sessionId); receiveError != nil {
			break
		}
		if receiveError = binary.Read(bytes.NewReader(raw[headerSize+4:headerSize+8]), binary.BigEndian, &length); receiveError != nil {
			break
		}
		payload, err := _readN(socketContext.conn, int(length))
		if err != nil {
			receiveError = err
			break
		}
		tail, err := _readN(socketContext.conn, len(TAILER_BYTES))
		if err != nil {
			receiveError = err
			break
		}
		if !bytes.Equal(tail, []byte(TAILER_BYTES)) {
			receiveError = fmt.Errorf("verify message tail failed on channel: %s", socketId)
			break
		}
		socketContext.writeLock.Lock()
		callback := socketContext.callbackMap[strconv.Itoa(int(sessionId))]
		socketContext.writeLock.Unlock()
		if callback != nil {
			callback.OnData(payload)
		} else {
			socketContext.queue.push(response{sessionId: sessionId, payload: payload})
		}
	}
}

func (uKtp *uitestKitTransport) sendMessage(reqType ReqTypeEnum, sessionId uint32, message string) (response UitestKitResponse, err error) {
	defer func() {
		if err != nil {
			uKtp.disconnect(reqType)
		}
	}()
	if err = uKtp._sendMessage(reqType, sessionId, message); err != nil {
		return
	}
	return uKtp.receiveMessage(reqType, sessionId)
}

func (uKtp *uitestKitTransport) _sendMessage(reqType ReqTypeEnum, sessionId uint32, message string) (err error) {
	socketId := fmt.Sprintf("socket_%d_%s", reqType, uKtp.serial)
	socketContext := uKtp.socketMap[socketId]
	if socketContext == nil {
		if err = uKtp.initializeSocket(reqType); err != nil {
			return
		}
		socketContext = uKtp.socketMap[socketId]
	}
	buffer := new(bytes.Buffer)
	if err = binary.Write(buffer, binary.BigEndian, []byte(HEADER_BYTES)); err != nil {
		return
	}
	if err = binary.Write(buffer, binary.BigEndian, sessionId); err != nil {
		return
	}
	if err = binary.Write(buffer, binary.BigEndian, uint32(len(message))); err != nil {
		return
	}
	if err = binary.Write(buffer, binary.BigEndian, []byte(message)); err != nil {
		return
	}
	if err = binary.Write(buffer, binary.BigEndian, []byte(TAILER_BYTES)); err != nil {
		return
	}
	socketContext.writeLock.Lock()
	defer socketContext.writeLock.Unlock()
	return _send(socketContext.conn, buffer.Bytes())
}

func (sc *SocketContext) Close() {
	sc.writeLock.Lock()
	defer sc.writeLock.Unlock()

	if sc.conn != nil {
		_ = sc.conn.Close()
	}

	// 清空callbackMap
	for key := range sc.callbackMap {
		delete(sc.callbackMap, key)
	}
	sc.callbackMap = nil

	// 清理队列
	if sc.queue != nil {
		sc.queue.Clear()
	}
}

func (uKtp *uitestKitTransport) Close() {
	uKtp.mu.Lock()
	defer uKtp.mu.Unlock()

	// 关闭所有的SocketContext
	if uKtp.socketMap != nil {
		for _, socketContext := range uKtp.socketMap {
			socketContext.Close()
		}
	}

	// 关闭连接池
	if uKtp.connectionPool != nil {
		uKtp.connectionPool.close()
	}

	uKtp.socketMap = nil
	uKtp.connectionPool = nil
}

func (uKtp *uitestKitTransport) receiveMessage(reqType ReqTypeEnum, sessionId uint32) (response UitestKitResponse, err error) {
	socketId := fmt.Sprintf("socket_%d_%s", reqType, uKtp.serial)
	socketContext := uKtp.socketMap[socketId]
	if socketContext == nil {
		err = fmt.Errorf("failed to read message. not found target connection")
		return
	}
	// 创建一个计时器，设置超时时间为 10 秒
	timeout := time.After(2 * time.Second)

	// 创建一个 ticker，每秒触发一次
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			err = fmt.Errorf("failed to read message in 2 second")
			return
		case <-ticker.C:
			res := socketContext.queue.remove(sessionId)
			if res != nil {
				err = json.Unmarshal(res.payload, &response)
				return
			}
		}
	}
}

func (sc *SocketContext) onException(err error) {
	for _, callback := range sc.callbackMap {
		if callback != nil {
			callback.OnError(err)
		}
	}
}

func newResponseList() *responseList {
	return &responseList{list: list.New()}
}
