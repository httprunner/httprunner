package gidevice

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"strconv"
	"time"

	"github.com/httprunner/httprunner/v4/hrp/pkg/gidevice/pkg/libimobiledevice"
)

type PerfOptions struct {
	// system
	SysCPU     bool `json:"sys_cpu,omitempty" yaml:"sys_cpu,omitempty"`
	SysMem     bool `json:"sys_mem,omitempty" yaml:"sys_mem,omitempty"`
	SysDisk    bool `json:"sys_disk,omitempty" yaml:"sys_disk,omitempty"`
	SysNetwork bool `json:"sys_network,omitempty" yaml:"sys_network,omitempty"`
	gpu        bool
	FPS        bool `json:"fps,omitempty" yaml:"fps,omitempty"`
	Network    bool `json:"network,omitempty" yaml:"network,omitempty"`
	// process
	BundleID string `json:"bundle_id,omitempty" yaml:"bundle_id,omitempty"`
	Pid      int    `json:"pid,omitempty" yaml:"pid,omitempty"`
	// config
	OutputInterval    int      `json:"output_interval,omitempty" yaml:"output_interval,omitempty"` // ms
	SystemAttributes  []string `json:"system_attributes,omitempty" yaml:"system_attributes,omitempty"`
	ProcessAttributes []string `json:"process_attributes,omitempty" yaml:"process_attributes,omitempty"`
}

func defaulPerfOption() *PerfOptions {
	return &PerfOptions{
		SysCPU:         false,
		SysMem:         false,
		SysDisk:        false,
		SysNetwork:     false,
		gpu:            false,
		FPS:            false,
		Network:        false,
		OutputInterval: 1000, // default 1000ms
		SystemAttributes: []string{
			// disk
			"diskBytesRead",
			"diskBytesWritten",
			"diskReadOps",
			"diskWriteOps",
			// memory
			"vmCompressorPageCount",
			"vmExtPageCount",
			"vmFreeCount",
			"vmIntPageCount",
			"vmPurgeableCount",
			"vmWireCount",
			"vmUsedCount",
			"__vmSwapUsage",
			// network
			"netBytesIn",
			"netBytesOut",
			"netPacketsIn",
			"netPacketsOut",
		},
		ProcessAttributes: []string{
			"pid",
			"cpuUsage",
		},
	}
}

type PerfOption func(*PerfOptions)

func WithPerfSystemCPU(b bool) PerfOption {
	return func(opt *PerfOptions) {
		opt.SysCPU = b
	}
}

func WithPerfSystemMem(b bool) PerfOption {
	return func(opt *PerfOptions) {
		opt.SysMem = b
	}
}

func WithPerfSystemDisk(b bool) PerfOption {
	return func(opt *PerfOptions) {
		opt.SysDisk = b
	}
}

func WithPerfSystemNetwork(b bool) PerfOption {
	return func(opt *PerfOptions) {
		opt.SysNetwork = b
	}
}

func WithPerfBundleID(bundleID string) PerfOption {
	return func(opt *PerfOptions) {
		opt.BundleID = bundleID
	}
}

func WithPerfPID(pid int) PerfOption {
	return func(opt *PerfOptions) {
		opt.Pid = pid
	}
}

func WithPerfGPU(b bool) PerfOption {
	return func(opt *PerfOptions) {
		opt.gpu = b
	}
}

func WithPerfFPS(b bool) PerfOption {
	return func(opt *PerfOptions) {
		opt.FPS = b
	}
}

func WithPerfNetwork(b bool) PerfOption {
	return func(opt *PerfOptions) {
		opt.Network = b
	}
}

func WithPerfOutputInterval(intervalMilliseconds int) PerfOption {
	return func(opt *PerfOptions) {
		opt.OutputInterval = intervalMilliseconds
	}
}

func WithPerfProcessAttributes(attrs ...string) PerfOption {
	return func(opt *PerfOptions) {
		opt.ProcessAttributes = attrs
	}
}

func WithPerfSystemAttributes(attrs ...string) PerfOption {
	return func(opt *PerfOptions) {
		opt.SystemAttributes = attrs
	}
}

type perfdClient struct {
	options *PerfOptions
	i       Instruments
	stop    chan struct{}      // used to stop perf client
	cancel  context.CancelFunc // used to cancel all iterators
}

func (d *device) newPerfdSysmontap(options *PerfOptions) (*perfdSysmontap, error) {
	instruments, err := d.newInstrumentsService()
	if err != nil {
		return nil, err
	}
	return &perfdSysmontap{
		perfdClient: perfdClient{
			i:       instruments,
			options: options,
			stop:    make(chan struct{}),
		},
		chanSysCPU:     make(chan []byte, 10),
		chanSysMem:     make(chan []byte, 10),
		chanSysDisk:    make(chan []byte, 10),
		chanSysNetwork: make(chan []byte, 10),
		chanProcess:    make(chan []byte, 10),
	}, nil
}

type perfdSysmontap struct {
	perfdClient
	chanSysCPU     chan []byte // system cpu channel
	chanSysMem     chan []byte // system mem channel
	chanSysDisk    chan []byte // system disk channel
	chanSysNetwork chan []byte // system network channel
	chanProcess    chan []byte // process channel
}

func (c *perfdSysmontap) Start() (data <-chan []byte, err error) {
	// set config
	interval := time.Millisecond * time.Duration(c.options.OutputInterval)
	log.Printf("set sysmontap sample interval: %dms\n", c.options.OutputInterval)

	config := map[string]interface{}{
		"bm":             0,
		"cpuUsage":       true,
		"sampleInterval": interval,                    // time.Duration
		"ur":             c.options.OutputInterval,    // 输出频率
		"procAttrs":      c.options.ProcessAttributes, // process performance
		"sysAttrs":       c.options.SystemAttributes,  // system performance
	}
	if _, err = c.i.call(
		instrumentsServiceSysmontap,
		"setConfig:",
		config,
	); err != nil {
		return nil, err
	}

	// start
	if _, err = c.i.call(
		instrumentsServiceSysmontap,
		"start",
	); err != nil {
		return nil, err
	}

	// register listener
	ctx, cancel := context.WithCancel(context.TODO())
	c.i.registerCallback("", func(m libimobiledevice.DTXMessageResult) {
		select {
		case <-ctx.Done():
			c.i.call(instrumentsServiceSysmontap, "stop")
			return
		default:
			dataArray, ok := m.Obj.([]interface{})
			if !ok || len(dataArray) != 2 {
				return
			}

			if c.options.Pid != 0 {
				c.parseProcessData(dataArray)
			} else {
				c.parseSystemData(dataArray)
			}
		}
	})
	c.cancel = cancel

	outCh := make(chan []byte, 100)

	go func() {
		for {
			select {
			case <-c.stop:
				c.cancel()
				return
			case cpuBytes, ok := <-c.chanSysCPU:
				if ok {
					outCh <- cpuBytes
				}
			case memBytes, ok := <-c.chanSysMem:
				if ok {
					outCh <- memBytes
				}
			case diskBytes, ok := <-c.chanSysDisk:
				if ok {
					outCh <- diskBytes
				}
			case networkBytes, ok := <-c.chanSysNetwork:
				if ok {
					outCh <- networkBytes
				}
			case processBytes, ok := <-c.chanProcess:
				if ok {
					outCh <- processBytes
				}
			}
		}
	}()

	return outCh, nil
}

func (c *perfdSysmontap) Stop() {
	close(c.stop)
}

func (c *perfdSysmontap) parseProcessData(dataArray []interface{}) {
	// dataArray example:
	// [
	//   map[
	//     CPUCount:2
	//     EnabledCPUs:2
	//     PerCPUUsage:[
	//       map[CPU_NiceLoad:0 CPU_SystemLoad:-1 CPU_TotalLoad:3.6363636363636402 CPU_UserLoad:-1]
	//       map[CPU_NiceLoad:0 CPU_SystemLoad:-1 CPU_TotalLoad:2.7272727272727195 CPU_UserLoad:-1]
	//     ]
	//     System:[36408520704 6897049600 3031160 773697 15596 61940 1297 26942 588 17020 127346 1835008 119718056 107009899 174046 103548]
	//     SystemCPUUsage:map[CPU_NiceLoad:0 CPU_SystemLoad:-1 CPU_TotalLoad:6.36363636363636 CPU_UserLoad:-1]
	//     StartMachAbsTime:5896602132889
	//     EndMachAbsTime:5896628486761
	//     Type:41
	//  ]
	//  map[
	//    Processes:map[
	//      0:[1.3582834340402803 0]
	//      124:[0.011456702068519481 124]
	//      136:[0.05468332721703649 136]
	//    ]
	//    StartMachAbsTime:5896602295095
	//    EndMachAbsTime:5896628780514
	//    Type:5
	//   ]
	// ]

	processData := make(map[string]interface{})
	processData["type"] = "process"
	processData["timestamp"] = time.Now().Unix()
	processData["pid"] = c.options.Pid

	defer func() {
		processBytes, _ := json.Marshal(processData)
		c.chanProcess <- processBytes
	}()

	systemInfo := dataArray[0].(map[string]interface{})
	processInfo := dataArray[1].(map[string]interface{})
	if _, ok := systemInfo["System"]; !ok {
		systemInfo, processInfo = processInfo, systemInfo
	}

	var targetProcessValue []interface{}
	processList := processInfo["Processes"].(map[string]interface{})
	for pid, v := range processList {
		if pid != strconv.Itoa(c.options.Pid) {
			continue
		}
		targetProcessValue = v.([]interface{})
	}

	if targetProcessValue == nil {
		processData["msg"] = fmt.Sprintf("process %d not found", c.options.Pid)
		return
	}

	processAttributesMap := make(map[string]interface{})
	for idx, value := range c.options.ProcessAttributes {
		processAttributesMap[value] = targetProcessValue[idx]
	}
	processData["proc_perf"] = processAttributesMap

	systemAttributesValue := systemInfo["System"].([]interface{})
	systemAttributesMap := make(map[string]int64)
	for idx, value := range c.options.SystemAttributes {
		systemAttributesMap[value] = convert2Int64(systemAttributesValue[idx])
	}
	processData["sys_perf"] = systemAttributesMap
}

func (c *perfdSysmontap) parseSystemData(dataArray []interface{}) {
	timestamp := time.Now().Unix()
	var systemInfo map[string]interface{}
	data1 := dataArray[0].(map[string]interface{})
	data2 := dataArray[1].(map[string]interface{})
	if _, ok := data1["SystemCPUUsage"]; ok {
		systemInfo = data1
	} else {
		systemInfo = data2
	}

	// systemInfo example:
	// map[
	//   CPUCount:2
	//   EnabledCPUs:2
	//   PerCPUUsage:[
	//     map[CPU_NiceLoad:0 CPU_SystemLoad:-1 CPU_TotalLoad:3.9215686274509807 CPU_UserLoad:-1]
	//     map[CPU_NiceLoad:0 CPU_SystemLoad:-1 CPU_TotalLoad:11.650485436893206 CPU_UserLoad:-1]]
	//   ]
	//   System:[704211 35486281728 6303789056 3001119 1001 11033 52668 1740 40022 2114 17310 126903 1835008 160323 107909856 95067 95808179]
	//   SystemCPUUsage:map[
	//     CPU_NiceLoad:0
	//     CPU_SystemLoad:-1
	//     CPU_TotalLoad:15.572054064344186
	//     CPU_UserLoad:-1
	//   ]
	//   StartMachAbsTime:5339240248449
	//   EndMachAbsTime:5339264441260
	//   Type:41
	// ]

	if c.options.SysCPU {
		sysCPUUsage := systemInfo["SystemCPUUsage"].(map[string]interface{})
		sysCPUInfo := SystemCPUData{
			PerfDataBase: PerfDataBase{
				Type:      "sys_cpu",
				TimeStamp: timestamp,
			},
			NiceLoad:   sysCPUUsage["CPU_NiceLoad"].(float64),
			SystemLoad: sysCPUUsage["CPU_SystemLoad"].(float64),
			TotalLoad:  sysCPUUsage["CPU_TotalLoad"].(float64),
			UserLoad:   sysCPUUsage["CPU_UserLoad"].(float64),
		}
		cpuBytes, _ := json.Marshal(sysCPUInfo)
		c.chanSysCPU <- cpuBytes
	}

	systemAttributesValue := systemInfo["System"].([]interface{})
	systemAttributesMap := make(map[string]int64)
	for idx, value := range c.options.SystemAttributes {
		systemAttributesMap[value] = convert2Int64(systemAttributesValue[idx])
	}

	if c.options.SysMem {
		kernelPageSize := int64(1) // why 16384 ?
		appMemory := (systemAttributesMap["vmIntPageCount"] - systemAttributesMap["vmPurgeableCount"]) * kernelPageSize
		cachedFiles := (systemAttributesMap["vmExtPageCount"] - systemAttributesMap["vmPurgeableCount"]) * kernelPageSize
		compressed := systemAttributesMap["vmCompressorPageCount"] * kernelPageSize
		usedMemory := (systemAttributesMap["vmUsedCount"] - systemAttributesMap["vmExtPageCount"]) * kernelPageSize
		wiredMemory := systemAttributesMap["vmWireCount"] * kernelPageSize
		swapUsed := systemAttributesMap["__vmSwapUsage"]
		freeMemory := systemAttributesMap["vmFreeCount"] * kernelPageSize

		sysMemInfo := SystemMemData{
			PerfDataBase: PerfDataBase{
				Type:      "sys_mem",
				TimeStamp: timestamp,
			},
			AppMemory:   appMemory,
			UsedMemory:  usedMemory,
			WiredMemory: wiredMemory,
			FreeMemory:  freeMemory,
			CachedFiles: cachedFiles,
			Compressed:  compressed,
			SwapUsed:    swapUsed,
		}
		memBytes, _ := json.Marshal(sysMemInfo)
		c.chanSysMem <- memBytes
	}

	if c.options.SysDisk {
		diskBytesRead := systemAttributesMap["diskBytesRead"]
		diskBytesWritten := systemAttributesMap["diskBytesWritten"]
		diskReadOps := systemAttributesMap["diskReadOps"]
		diskWriteOps := systemAttributesMap["diskWriteOps"]

		sysDiskInfo := SystemDiskData{
			PerfDataBase: PerfDataBase{
				Type:      "sys_disk",
				TimeStamp: timestamp,
			},
			DataRead:    diskBytesRead,
			DataWritten: diskBytesWritten,
			ReadOps:     diskReadOps,
			WriteOps:    diskWriteOps,
		}
		diskBytes, _ := json.Marshal(sysDiskInfo)
		c.chanSysDisk <- diskBytes
	}

	if c.options.SysNetwork {
		netBytesIn := systemAttributesMap["netBytesIn"]
		netBytesOut := systemAttributesMap["netBytesOut"]
		netPacketsIn := systemAttributesMap["netPacketsIn"]
		netPacketsOut := systemAttributesMap["netPacketsOut"]

		sysNetworkInfo := SystemNetworkData{
			PerfDataBase: PerfDataBase{
				Type:      "sys_network",
				TimeStamp: timestamp,
			},
			BytesIn:    netBytesIn,
			BytesOut:   netBytesOut,
			PacketsIn:  netPacketsIn,
			PacketsOut: netPacketsOut,
		}
		networkBytes, _ := json.Marshal(sysNetworkInfo)
		c.chanSysNetwork <- networkBytes
	}
}

type SystemCPUData struct {
	PerfDataBase         // system cpu
	NiceLoad     float64 `json:"nice_load"`
	SystemLoad   float64 `json:"system_load"`
	TotalLoad    float64 `json:"total_load"`
	UserLoad     float64 `json:"user_load"`
}

type SystemMemData struct {
	PerfDataBase       // mem
	AppMemory    int64 `json:"app_memory"`
	FreeMemory   int64 `json:"free_memory"`
	UsedMemory   int64 `json:"used_memory"`
	WiredMemory  int64 `json:"wired_memory"`
	CachedFiles  int64 `json:"cached_files"`
	Compressed   int64 `json:"compressed"`
	SwapUsed     int64 `json:"swap_used"`
}

type SystemDiskData struct {
	PerfDataBase       // disk
	DataRead     int64 `json:"data_read"`
	DataWritten  int64 `json:"data_written"`
	ReadOps      int64 `json:"reads_in"`
	WriteOps     int64 `json:"writes_out"`
}

type SystemNetworkData struct {
	PerfDataBase       // network
	BytesIn      int64 `json:"bytes_in"`
	BytesOut     int64 `json:"bytes_out"`
	PacketsIn    int64 `json:"packets_in"`
	PacketsOut   int64 `json:"packets_out"`
}

func (d *device) newPerfdNetworking(options *PerfOptions) (*perfdNetworking, error) {
	instruments, err := d.newInstrumentsService()
	if err != nil {
		return nil, err
	}
	return &perfdNetworking{
		perfdClient: perfdClient{
			i:       instruments,
			options: options,
			stop:    make(chan struct{}),
		},
		chanNetwork: make(chan []byte, 10),
	}, nil
}

type perfdNetworking struct {
	perfdClient
	chanNetwork chan []byte // network channel
}

func (c *perfdNetworking) Start() (data <-chan []byte, err error) {
	if _, err = c.i.call(
		instrumentsServiceNetworking,
		"replayLastRecordedSession",
	); err != nil {
		return nil, err
	}

	if _, err = c.i.call(
		instrumentsServiceNetworking,
		"startMonitoring",
	); err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.TODO())
	c.i.registerCallback("", func(m libimobiledevice.DTXMessageResult) {
		select {
		case <-ctx.Done():
			c.i.call(instrumentsServiceNetworking, "stopMonitoring")
			return
		default:
			c.parseNetworking(m.Obj)
		}
	})
	c.cancel = cancel

	outCh := make(chan []byte, 100)

	go func() {
		for {
			select {
			case <-c.stop:
				c.cancel()
				return
			case networkBytes, ok := <-c.chanNetwork:
				if ok {
					outCh <- networkBytes
				}
			}
		}
	}()

	return outCh, nil
}

func (c *perfdNetworking) Stop() {
	close(c.stop)
}

func (c *perfdNetworking) parseNetworking(data interface{}) {
	raw, ok := data.([]interface{})
	if !ok || len(raw) != 2 {
		fmt.Printf("invalid networking data: %v\n", data)
		return
	}

	var netBytes []byte
	msgType := raw[0].(uint64)
	msgValue := raw[1].([]interface{})
	if msgType == 0 {
		// interface-detection
		// ['InterfaceIndex', "Name"]
		// e.g. [0, [14, 'en0']]
		netData := NetworkDataInterfaceDetection{
			PerfDataBase: PerfDataBase{
				Type:      "network-interface-detection",
				TimeStamp: time.Now().Unix(),
			},
			InterfaceIndex: convert2Int64(msgValue[0]),
			Name:           msgValue[1].(string),
		}
		netBytes, _ = json.Marshal(netData)
	} else if msgType == 1 {
		// connection-detected
		// ['LocalAddress', 'RemoteAddress', 'InterfaceIndex', 'Pid',
		// 'RecvBufferSize', 'RecvBufferUsed', 'SerialNumber', 'Kind']
		// e.g. [1 [[16 2 211 158 192 168 100 101 0 0 0 0 0 0 0 0]
		//       [16 2 0 53 183 221 253 100 0 0 0 0 0 0 0 0]
		//       14 -2 786896 0 133 2]]

		localAddr, err := parseSocketAddr(msgValue[0].([]byte))
		if err != nil {
			fmt.Printf("parse local socket address err: %v\n", err)
		}
		remoteAddr, err := parseSocketAddr(msgValue[1].([]byte))
		if err != nil {
			fmt.Printf("parse remote socket address err: %v\n", err)
		}
		netData := NetworkDataConnectionDetected{
			PerfDataBase: PerfDataBase{
				Type:      "network-connection-detected",
				TimeStamp: time.Now().Unix(),
			},
			LocalAddress:   localAddr,
			RemoteAddress:  remoteAddr,
			InterfaceIndex: convert2Int64(msgValue[2]),
			Pid:            convert2Int64(msgValue[3]),
			RecvBufferSize: convert2Int64(msgValue[4]),
			RecvBufferUsed: convert2Int64(msgValue[5]),
			SerialNumber:   convert2Int64(msgValue[6]),
			Kind:           convert2Int64(msgValue[7]),
		}
		netBytes, _ = json.Marshal(netData)
	} else if msgType == 2 {
		// connection-update
		// ['RxPackets', 'RxBytes', 'TxPackets', 'TxBytes',
		// 'RxDups', 'RxOOO', 'TxRetx', 'MinRTT', 'AvgRTT', 'ConnectionSerial']
		// e.g. [2, [21, 1708, 22, 14119, 309, 0, 5830, 0.076125, 0.076125, 54, -1]]
		netData := NetworkDataConnectionUpdate{
			PerfDataBase: PerfDataBase{
				Type:      "network-connection-update",
				TimeStamp: time.Now().Unix(),
			},
			RxBytes:   convert2Int64(msgValue[0]),
			RxPackets: convert2Int64(msgValue[1]),
			TxBytes:   convert2Int64(msgValue[2]),
			TxPackets: convert2Int64(msgValue[3]),
		}
		if value, ok := msgValue[4].(uint64); ok {
			netData.RxDups = int64(value)
		}
		if value, ok := msgValue[5].(uint64); ok {
			netData.RxOOO = int64(value)
		}
		if value, ok := msgValue[6].(uint64); ok {
			netData.TxRetx = int64(value)
		}
		if value, ok := msgValue[7].(uint64); ok {
			netData.MinRTT = int64(value)
		}
		if value, ok := msgValue[8].(uint64); ok {
			netData.AvgRTT = int64(value)
		}
		if value, ok := msgValue[9].(uint64); ok {
			netData.ConnectionSerial = int64(value)
		}

		netBytes, _ = json.Marshal(netData)
	}
	c.chanNetwork <- netBytes
}

func parseSocketAddr(data []byte) (string, error) {
	len := data[0]                             // length of address
	_ = data[1]                                // family
	port := binary.BigEndian.Uint16(data[2:4]) // port

	// network, data[4:4+len]
	if len == 0x10 {
		// IPv4, 4 bytes
		ip := net.IP(data[4:8])
		return fmt.Sprintf("%s:%d", ip, port), nil
	} else if len == 0x1c {
		// IPv6, 16 bytes
		ip := net.IP(data[4:20])
		return fmt.Sprintf("%s:%d", ip, port), nil
	}
	return "", fmt.Errorf("invalid socket address: %v", data)
}

type PerfDataBase struct {
	Type      string `json:"type"`
	TimeStamp int64  `json:"timestamp"`
	Msg       string `json:"msg,omitempty"` // message for invalid data
}

// network-interface-detection
type NetworkDataInterfaceDetection struct {
	PerfDataBase
	InterfaceIndex int64  `json:"interface_index"` // 0
	Name           string `json:"name"`            // 1
}

// network-connection-detected
type NetworkDataConnectionDetected struct {
	PerfDataBase
	LocalAddress   string `json:"local_address"`    // 0
	RemoteAddress  string `json:"remote_address"`   // 1
	InterfaceIndex int64  `json:"interface_index"`  // 2
	Pid            int64  `json:"pid"`              // 3
	RecvBufferSize int64  `json:"recv_buffer_size"` // 4
	RecvBufferUsed int64  `json:"recv_buffer_used"` // 5
	SerialNumber   int64  `json:"serial_number"`    // 6
	Kind           int64  `json:"kind"`             // 7
}

// network-connection-update
type NetworkDataConnectionUpdate struct {
	PerfDataBase
	RxBytes          int64 `json:"rx_bytes"`          // 0
	RxPackets        int64 `json:"rx_packets"`        // 1
	TxBytes          int64 `json:"tx_bytes"`          // 2
	TxPackets        int64 `json:"tx_packets"`        // 3
	RxDups           int64 `json:"rx_dups,omitempty"` // 4
	RxOOO            int64 `json:"rx_000,omitempty"`  // 5
	TxRetx           int64 `json:"tx_retx,omitempty"` // 6
	MinRTT           int64 `json:"min_rtt,omitempty"` // 7
	AvgRTT           int64 `json:"avg_rtt,omitempty"` // 8
	ConnectionSerial int64 `json:"connection_serial"` // 9
}

func (d *device) newPerfdGraphicsOpengl(options *PerfOptions) (*perfdGraphicsOpengl, error) {
	instruments, err := d.newInstrumentsService()
	if err != nil {
		return nil, err
	}
	return &perfdGraphicsOpengl{
		perfdClient: perfdClient{
			i:       instruments,
			options: options,
			stop:    make(chan struct{}),
		},
		chanGPU: make(chan []byte, 10),
		chanFPS: make(chan []byte, 10),
	}, nil
}

type perfdGraphicsOpengl struct {
	perfdClient
	chanGPU chan []byte // gpu channel
	chanFPS chan []byte // fps channel
}

func (c *perfdGraphicsOpengl) Start() (data <-chan []byte, err error) {
	if _, err = c.i.call(
		instrumentsServiceGraphicsOpengl,
		"setSamplingRate:",
		float64(c.options.OutputInterval)/100, // FIXME: unable to set sampling rate, always 1.0
	); err != nil {
		return nil, err
	}

	if _, err = c.i.call(
		instrumentsServiceGraphicsOpengl,
		"startSamplingAtTimeInterval:",
		0,
	); err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.TODO())
	c.i.registerCallback("", func(m libimobiledevice.DTXMessageResult) {
		select {
		case <-ctx.Done():
			c.i.call(instrumentsServiceGraphicsOpengl, "stopSampling")
			return
		default:
			c.parseData(m.Obj)
		}
	})
	c.cancel = cancel

	outCh := make(chan []byte, 100)

	go func() {
		for {
			select {
			case <-c.stop:
				c.cancel()
				return
			case gpuBytes, ok := <-c.chanGPU:
				if ok {
					outCh <- gpuBytes
				}
			case fpsBytes, ok := <-c.chanFPS:
				if ok {
					outCh <- fpsBytes
				}
			}
		}
	}()

	return outCh, nil
}

func (c *perfdGraphicsOpengl) Stop() {
	close(c.stop)
}

func (c *perfdGraphicsOpengl) parseData(data interface{}) {
	// data example:
	// map[
	//   Alloc system memory:50167808
	//   Allocated PB Size:1179648
	//   CoreAnimationFramesPerSecond:0  // fps from GPU
	//   Device Utilization %:0          // device
	//   IOGLBundleName:Built-In
	//   In use system memory:10633216
	//   Renderer Utilization %:0        // renderer
	//   SplitSceneCount:0
	//   TiledSceneBytes:0
	//   Tiler Utilization %:0           // tiler
	//   XRVideoCardRunTimeStamp:1010679
	//   recoveryCount:0
	// ]

	gpuInfo := GPUData{
		PerfDataBase: PerfDataBase{
			Type:      "gpu",
			TimeStamp: time.Now().Unix(),
		},
	}
	fpsInfo := FPSData{
		PerfDataBase: PerfDataBase{
			Type:      "fps",
			TimeStamp: time.Now().Unix(),
		},
	}

	defer func() {
		if c.options.gpu {
			gpuBytes, _ := json.Marshal(gpuInfo)
			c.chanGPU <- gpuBytes
		}
		if c.options.FPS {
			fpsBytes, _ := json.Marshal(fpsInfo)
			c.chanFPS <- fpsBytes
		}
	}()

	raw, ok := data.(map[string]interface{})
	if !ok {
		gpuInfo.Msg = fmt.Sprintf("invalid graphics.opengl data: %v", data)
		return
	}

	// gpu
	gpuInfo.DeviceUtilization = convert2Int64(raw["Device Utilization %"])
	gpuInfo.TilerUtilization = convert2Int64(raw["Tiler Utilization %"])
	gpuInfo.RendererUtilization = convert2Int64(raw["Renderer Utilization %"])

	// fps
	fpsInfo.FPS = int(convert2Int64(raw["CoreAnimationFramesPerSecond"]))
}

type GPUData struct {
	PerfDataBase              // gpu
	TilerUtilization    int64 `json:"tiler_utilization"`    // 处理顶点的 GPU 时间占比
	DeviceUtilization   int64 `json:"device_utilization"`   // 设备利用率
	RendererUtilization int64 `json:"renderer_utilization"` // 渲染器利用率
}

type FPSData struct {
	PerfDataBase     // fps
	FPS          int `json:"fps"`
}

func convert2Int64(num interface{}) int64 {
	switch value := num.(type) {
	case int64:
		return value
	case uint64:
		return int64(value)
	case uint32:
		return int64(value)
	case uint16:
		return int64(value)
	case uint8:
		return int64(value)
	case uint:
		return int64(value)
	}
	fmt.Printf("convert2Int64 failed: %v, %T\n", num, num)
	return -1
}

func containString(ss []string, s string) bool {
	for _, v := range ss {
		if s == v {
			return true
		}
	}
	return false
}
