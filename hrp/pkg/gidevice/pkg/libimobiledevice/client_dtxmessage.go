package libimobiledevice

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"
	"unsafe"

	"github.com/httprunner/httprunner/v4/hrp/pkg/gidevice/pkg/nskeyedarchiver"
)

const (
	_unregistered = "_Golang-iDevice_Unregistered"
	_over         = "_Golang-iDevice_Over"
)

func newDtxMessageClient(innerConn InnerConn) *dtxMessageClient {
	c := &dtxMessageClient{
		innerConn:         innerConn,
		msgID:             0,
		publishedChannels: make(map[string]int32),
		openedChannels:    make(map[string]uint32),
		toReply:           make(chan *dtxMessageHeaderPacket),

		mu:        sync.Mutex{},
		resultMap: make(map[interface{}]*DTXMessageResult),

		callbackMap: make(map[string]func(m DTXMessageResult)),
	}
	c.RegisterCallback(_unregistered, func(m DTXMessageResult) {})
	c.RegisterCallback(_over, func(m DTXMessageResult) {})
	c.ctx, c.cancelFunc = context.WithCancel(context.Background())
	c.startReceive()
	c.startWaitingForReply()
	return c
}

type dtxMessageClient struct {
	innerConn InnerConn
	msgID     uint32

	publishedChannels map[string]int32
	openedChannels    map[string]uint32

	toReply chan *dtxMessageHeaderPacket

	mu        sync.Mutex
	resultMap map[interface{}]*DTXMessageResult

	callbackMap map[string]func(m DTXMessageResult)

	ctx        context.Context
	cancelFunc context.CancelFunc
}

func (c *dtxMessageClient) SendDTXMessage(selector string, aux []byte, channelCode uint32, expectsReply bool) (msgID uint32, err error) {
	payload := new(dtxMessagePayloadPacket)
	header := &dtxMessageHeaderPacket{
		ExpectsReply: 1,
	}

	flag := 0x1000
	if !expectsReply {
		flag = 0
		header.ExpectsReply = 0
	}

	var sel []byte
	if sel, err = nskeyedarchiver.Marshal(selector); err != nil {
		return 0, err
	}

	if aux == nil {
		aux = make([]byte, 0)
	}

	payload.Flags = uint32(0x2 | flag)
	payload.AuxiliaryLength = uint32(len(aux))
	payload.TotalLength = uint64(len(aux)) + uint64(len(sel))

	header.Magic = 0x1F3D5B79
	header.CB = uint32(unsafe.Sizeof(*header))
	header.FragmentId = 0
	header.FragmentCount = 1
	header.Length = uint32(unsafe.Sizeof(*payload)) + uint32(payload.TotalLength)
	c.msgID++
	header.Identifier = c.msgID
	header.ConversationIndex = 0
	header.ChannelCode = channelCode

	msgPkt := new(dtxMessagePacket)
	msgPkt.Header = header
	msgPkt.Payload = payload
	msgPkt.Aux = aux
	msgPkt.Sel = sel

	raw, err := msgPkt.Pack()
	if err != nil {
		return 0, err
	}

	debugLog(fmt.Sprintf("--> %s\n", msgPkt))
	msgID = header.Identifier
	err = c.innerConn.Write(raw)
	return
}

func (c *dtxMessageClient) ReceiveDTXMessage() (result *DTXMessageResult, err error) {
	bufPayload := new(bytes.Buffer)

	var header *dtxMessageHeaderPacket = nil
	var needToReply *dtxMessageHeaderPacket = nil

	for {
		header = new(dtxMessageHeaderPacket)

		lenHeader := int(unsafe.Sizeof(*header))
		var bufHeader []byte
		if bufHeader, err = c.innerConn.Read(lenHeader); err != nil {
			return nil, fmt.Errorf("receive: length of DTXMessageHeader: %w", err)
		}

		if header, err = header.unpack(bytes.NewBuffer(bufHeader)); err != nil {
			return nil, fmt.Errorf("receive: DTXMessageHeader unpack: %w", err)
		}

		if header.ExpectsReply == 1 {
			needToReply = header
		}

		if header.Magic != 0x1F3D5B79 {
			return nil, fmt.Errorf("receive: bad magic %x", header.Magic)
		}

		if header.ConversationIndex == 1 {
			if header.Identifier != c.msgID {
				return nil, fmt.Errorf("receive: except identifier %d new identifier %d", c.msgID, header.Identifier)
			}
		} else if header.ConversationIndex == 0 {
			if header.Identifier > c.msgID {
				c.msgID = header.Identifier
			}
		} else {
			return nil, fmt.Errorf("receive: invalid conversationIndex %d", header.ConversationIndex)
		}

		if header.FragmentId == 0 && header.FragmentCount > 1 {
			continue
		}

		var data []byte
		if data, err = c.innerConn.Read(int(header.Length)); err != nil {
			return nil, fmt.Errorf("receive: length of DTXMessageHeader: %w", err)
		}
		bufPayload.Write(data)

		if header.FragmentId == header.FragmentCount-1 {
			break
		}
	}

	rawPayload := bufPayload.Bytes()
	payload := new(dtxMessagePayloadPacket)
	if payload, err = payload.unpack(bufPayload); err != nil {
		return nil, fmt.Errorf("receive: unpack DTXMessagePayload: %w", err)
	}

	compress := (payload.Flags & 0xff000) >> 12
	if compress != 0 {
		return nil, fmt.Errorf("receive: message is compressed type %d", compress)
	}

	payloadSize := uint32(unsafe.Sizeof(*payload))
	objOffset := uint64(payloadSize + payload.AuxiliaryLength)

	var aux, obj []byte

	// see https://github.com/electricbubble/gidevice/issues/28
	if r, l := payloadSize+payload.AuxiliaryLength, len(rawPayload); int(r) <= l {
		aux = rawPayload[payloadSize:r]
	} else {
		debugLog(fmt.Sprintf("<-- DTXMessage %s\n%s\n"+
			"[aux] bounds out of range [:%d] with capacity %d",
			header.String(), payload.String(),
			r, l,
		))
	}
	if r, l := objOffset+(payload.TotalLength-uint64(payload.AuxiliaryLength)), len(rawPayload); int(r) <= l {
		obj = rawPayload[objOffset:r]
	} else {
		debugLog(fmt.Sprintf("<-- DTXMessage %s\n%s\n"+
			"[obj] bounds out of range [:%d] with capacity %d",
			header.String(), payload.String(),
			r, l,
		))
	}

	debugLog(fmt.Sprintf(
		"<-- DTXMessage %s\n%s\n"+
			"%s\n%s\n",
		header.String(), payload.String(),
		hex.Dump(aux), hex.Dump(obj),
	))

	result = new(DTXMessageResult)

	if len(aux) > 0 {
		if aux, err := UnmarshalAuxBuffer(aux); err != nil {
			return nil, fmt.Errorf("receive: unpack AUX: %w", err)
		} else {
			result.Aux = aux
		}
	}

	if len(obj) > 0 {
		if obj, err := NewNSKeyedArchiver().Unmarshal(obj); err != nil {
			return nil, fmt.Errorf("receive: unpack NSKeyedArchiver: %w", err)
		} else {
			result.Obj = obj
		}
	}

	sObj, ok := result.Obj.(string)
	if fn, do := c.callbackMap[sObj]; do {
		fn(*result)
	} else {
		c.callbackMap[_unregistered](*result)
	}

	if needToReply != nil {
		go func() { c.toReply <- needToReply }()
	} else {
		var sk interface{} = header.Identifier

		if ok && sObj == "_notifyOfPublishedCapabilities:" {
			sk = "_notifyOfPublishedCapabilities:"
		}
		c.mu.Lock()
		c.resultMap[sk] = result
		c.mu.Unlock()
	}

	return
}

func (c *dtxMessageClient) Connection() (publishedChannels map[string]int32, err error) {
	args := NewAuxBuffer()
	if err = args.AppendObject(map[string]interface{}{
		"com.apple.private.DTXBlockCompression": uint64(2),
		"com.apple.private.DTXConnection":       uint64(1),
	}); err != nil {
		return nil, fmt.Errorf("connection DTXMessage: %w", err)
	}

	selector := "_notifyOfPublishedCapabilities:"
	if _, err = c.SendDTXMessage(selector, args.Bytes(), 0, false); err != nil {
		return nil, fmt.Errorf("connection send: %w", err)
	}

	var result *DTXMessageResult
	if result, err = c.GetResult(selector); err != nil {
		return nil, fmt.Errorf("connection receive: %w", err)
	}

	if result.Obj.(string) != "_notifyOfPublishedCapabilities:" {
		return nil, fmt.Errorf("connection: response mismatch: %s", result.Obj)
	}

	aux := result.Aux[0].(map[string]interface{})
	for k, v := range aux {
		c.publishedChannels[k] = int32(v.(uint64))
	}

	return c.publishedChannels, nil
}

func (c *dtxMessageClient) MakeChannel(channel string) (id uint32, err error) {
	var ok bool
	if id, ok = c.openedChannels[channel]; ok {
		return id, nil
	}

	id = uint32(len(c.openedChannels) + 1)
	args := NewAuxBuffer()
	args.AppendInt32(int32(id))
	if err = args.AppendObject(channel); err != nil {
		return 0, fmt.Errorf("make channel DTXMessage: %w", err)
	}

	selector := "_requestChannelWithCode:identifier:"

	var msgID uint32
	if msgID, err = c.SendDTXMessage(selector, args.Bytes(), 0, true); err != nil {
		return 0, fmt.Errorf("make channel send: %w", err)
	}

	if _, err = c.GetResult(msgID); err != nil {
		return 0, fmt.Errorf("make channel receive: %w", err)
	}

	c.openedChannels[channel] = id

	return
}

func (c *dtxMessageClient) RegisterCallback(obj string, cb func(m DTXMessageResult)) {
	c.callbackMap[obj] = cb
}

func (c *dtxMessageClient) GetResult(key interface{}) (*DTXMessageResult, error) {
	startTime := time.Now()
	for {
		time.Sleep(100 * time.Millisecond)
		c.mu.Lock()
		if v, ok := c.resultMap[key]; ok {
			delete(c.resultMap, key)
			c.mu.Unlock()
			return v, nil
		} else {
			c.mu.Unlock()
		}
		if elapsed := time.Since(startTime); elapsed > 30*time.Second {
			return nil, fmt.Errorf("dtx: get result: timeout after %v", elapsed)
		}
	}
}

func (c *dtxMessageClient) Close() {
	c.cancelFunc()
	c.innerConn.Close()
}

func (c *dtxMessageClient) startReceive() {
	go func() {
		for {
			select {
			case <-c.ctx.Done():
				return
			default:
				if _, err := c.ReceiveDTXMessage(); err != nil {
					debugLog(fmt.Sprintf("dtx: receive: %s", err))
					if strings.Contains(err.Error(), io.EOF.Error()) {
						c.cancelFunc()
						c.callbackMap[_over](DTXMessageResult{})
						break
					}
				}
			}
		}
	}()
}

func (c *dtxMessageClient) startWaitingForReply() {
	go func() {
		for {
			select {
			case <-c.ctx.Done():
				return
			case reqHeader := <-c.toReply:
				replyPayload := new(dtxMessagePayloadPacket)
				replyPayload.Flags = 0
				replyPayload.AuxiliaryLength = 0
				replyPayload.TotalLength = 0

				replyHeader := new(dtxMessageHeaderPacket)
				replyHeader.Magic = 0x1F3D5B79
				replyHeader.CB = uint32(unsafe.Sizeof(*replyHeader))
				replyHeader.FragmentId = 0
				replyHeader.FragmentCount = 1
				replyHeader.Length = uint32(unsafe.Sizeof(*replyPayload)) + uint32(replyPayload.TotalLength)
				replyHeader.Identifier = reqHeader.Identifier
				replyHeader.ConversationIndex = reqHeader.ConversationIndex + 1
				replyHeader.ChannelCode = reqHeader.ChannelCode
				replyHeader.ExpectsReply = 0

				replyPkt := new(dtxMessagePacket)
				replyPkt.Header = replyHeader
				replyPkt.Payload = replyPayload
				replyPkt.Aux = nil
				replyPkt.Sel = nil

				raw, err := replyPkt.Pack()
				if err != nil {
					debugLog(fmt.Sprintf("pack: reply DTXMessage: %s", err))
					continue
				}

				if err = c.innerConn.Write(raw); err != nil {
					debugLog(fmt.Sprintf("send: reply DTXMessage: %s", err))
					continue
				}
			}
		}
	}()
}

type DTXMessageResult struct {
	Obj interface{}
	Aux []interface{}
}
