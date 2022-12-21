package gidevice

import (
	"bytes"
	"context"
	"fmt"
	"time"

	"github.com/httprunner/httprunner/v4/hrp/pkg/gidevice/pkg/libimobiledevice"
	"github.com/httprunner/httprunner/v4/hrp/pkg/gidevice/pkg/nskeyedarchiver"
)

type Usbmux interface {
	Devices() ([]Device, error)
	ReadBUID() (string, error)
	Listen(chan Device) (context.CancelFunc, error)
}

type Device interface {
	Properties() DeviceProperties

	NewConnect(port int, timeout ...time.Duration) (InnerConn, error)
	ReadPairRecord() (pairRecord *PairRecord, err error)
	SavePairRecord(pairRecord *PairRecord) (err error)
	DeletePairRecord() (err error)

	lockdownService() (lockdown Lockdown, err error)
	QueryType() (LockdownType, error)
	GetValue(domain, key string) (v interface{}, err error)
	Pair() (pairRecord *PairRecord, err error)

	imageMounterService() (imageMounter ImageMounter, err error)
	Images(imgType ...string) (imageSignatures [][]byte, err error)
	MountDeveloperDiskImage(dmgPath string, signaturePath string) (err error)

	screenshotService() (lockdown Screenshot, err error)
	Screenshot() (raw *bytes.Buffer, err error)

	simulateLocationService() (simulateLocation SimulateLocation, err error)
	SimulateLocationUpdate(longitude float64, latitude float64, coordinateSystem ...CoordinateSystem) (err error)
	SimulateLocationRecover() (err error)

	installationProxyService() (installationProxy InstallationProxy, err error)
	InstallationProxyBrowse(opts ...InstallationProxyOption) (currentList []interface{}, err error)
	InstallationProxyLookup(opts ...InstallationProxyOption) (lookupResult interface{}, err error)

	instrumentsService() (instruments Instruments, err error)
	AppLaunch(bundleID string, opts ...AppLaunchOption) (pid int, err error)
	AppKill(pid int) (err error)
	AppRunningProcesses() (processes []Process, err error)
	AppList(opts ...AppListOption) (apps []Application, err error)
	DeviceInfo() (devInfo *DeviceInfo, err error)

	AfcService() (afc Afc, err error)
	AppInstall(ipaPath string) (err error)
	AppUninstall(bundleID string) (err error)

	HouseArrestService() (houseArrest HouseArrest, err error)

	syslogRelayService() (syslogRelay SyslogRelay, err error)
	Syslog() (lines <-chan string, err error)
	SyslogStop()

	PcapStart(opts ...PcapOption) (packet <-chan []byte, err error)
	PcapStop()

	Reboot() error
	Shutdown() error

	crashReportMoverService() (crashReportMover CrashReportMover, err error)
	MoveCrashReport(hostDir string, opts ...CrashReportMoverOption) (err error)

	XCTest(bundleID string, opts ...XCTestOption) (out <-chan string, cancel context.CancelFunc, err error)

	springBoardService() (springBoard SpringBoard, err error)
	GetIconPNGData(bundleId string) (raw *bytes.Buffer, err error)
	GetInterfaceOrientation() (orientation OrientationState, err error)

	PerfStart(opts ...PerfOption) (data <-chan []byte, err error)
	PerfStop()
}

type DeviceProperties = libimobiledevice.DeviceProperties

type OrientationState = libimobiledevice.OrientationState

type Lockdown interface {
	QueryType() (LockdownType, error)
	GetValue(domain, key string) (v interface{}, err error)
	SetValue(domain, key string, value interface{}) (err error)
	Pair() (pairRecord *PairRecord, err error)
	EnterRecovery() (err error)

	handshake() (err error)

	startSession(pairRecord *PairRecord) (err error)
	stopSession() (err error)
	startService(service string, escrowBag []byte) (dynamicPort int, enableSSL bool, err error)

	ImageMounterService() (imageMounter ImageMounter, err error)
	ScreenshotService() (screenshot Screenshot, err error)
	SimulateLocationService() (simulateLocation SimulateLocation, err error)
	InstallationProxyService() (installationProxy InstallationProxy, err error)
	InstrumentsService() (instruments Instruments, err error)
	TestmanagerdService() (testmanagerd Testmanagerd, err error)
	AfcService() (afc Afc, err error)
	HouseArrestService() (houseArrest HouseArrest, err error)
	SyslogRelayService() (syslogRelay SyslogRelay, err error)
	DiagnosticsRelayService() (diagnostics DiagnosticsRelay, err error)
	CrashReportMoverService() (crashReportMover CrashReportMover, err error)
	SpringBoardService() (springBoard SpringBoard, err error)
}

type ImageMounter interface {
	Images(imgType string) (imageSignatures [][]byte, err error)
	UploadImage(imgType, dmgPath string, signatureData []byte) (err error)
	Mount(imgType, devImgPath string, signatureData []byte) (err error)

	UploadImageAndMount(imgType, devImgPath, dmgPath, signaturePath string) (err error)
}

type Screenshot interface {
	exchange() (err error)
	Take() (raw *bytes.Buffer, err error)
}

type SimulateLocation interface {
	Update(longitude float64, latitude float64, coordinateSystem ...CoordinateSystem) (err error)
	// Recover try to revert back
	Recover() (err error)
}

type InstallationProxy interface {
	Browse(opts ...InstallationProxyOption) (currentList []interface{}, err error)
	Lookup(opts ...InstallationProxyOption) (lookupResult interface{}, err error)
	Install(bundleID, packagePath string) (err error)
	Uninstall(bundleID string) (err error)
}

type Instruments interface {
	AppLaunch(bundleID string, opts ...AppLaunchOption) (pid int, err error)
	AppKill(pid int) (err error)
	AppRunningProcesses() (processes []Process, err error)
	AppList(opts ...AppListOption) (apps []Application, err error)
	DeviceInfo() (devInfo *DeviceInfo, err error)

	getPidByBundleID(bundleID string) (pid int, err error)
	appProcess(bundleID string) (err error)
	startObserving(pid int) (err error)

	notifyOfPublishedCapabilities() (err error)
	requestChannel(channel string) (id uint32, err error)
	call(channel, selector string, auxiliaries ...interface{}) (result *libimobiledevice.DTXMessageResult, err error)

	// sysMonSetConfig(cfg ...interface{}) (err error)
	// SysMonStart(cfg ...interface{}) (_ interface{}, err error)

	registerCallback(obj string, cb func(m libimobiledevice.DTXMessageResult))
}

type Testmanagerd interface {
	notifyOfPublishedCapabilities() (err error)
	requestChannel(channel string) (id uint32, err error)
	newXCTestManagerDaemon() (xcTestManager XCTestManagerDaemon, err error)

	invoke(selector string, args *libimobiledevice.AuxBuffer, channel uint32, expectsReply bool) (*libimobiledevice.DTXMessageResult, error)

	registerCallback(obj string, cb func(m libimobiledevice.DTXMessageResult))
	close()
}

type Afc interface {
	DiskInfo() (diskInfo *AfcDiskInfo, err error)
	ReadDir(dirname string) (names []string, err error)
	Stat(filename string) (info *AfcFileInfo, err error)
	Open(filename string, mode AfcFileMode) (file *AfcFile, err error)
	Remove(filePath string) (err error)
	Rename(oldPath string, newPath string) (err error)
	Mkdir(path string) (err error)
	Link(oldName string, newName string, linkType AfcLinkType) (err error)
	Truncate(filePath string, size int64) (err error)
	SetFileModTime(filePath string, modTime time.Time) (err error)
	// Hash sha1 algorithm
	Hash(filePath string) ([]byte, error)
	// HashWithRange sha1 algorithm with file range
	HashWithRange(filePath string, start, end uint64) ([]byte, error)
	RemoveAll(path string) (err error)

	WriteFile(filename string, data []byte, perm AfcFileMode) (err error)
}

type HouseArrest interface {
	Documents(bundleID string) (afc Afc, err error)
	Container(bundleID string) (afc Afc, err error)
}

type XCTestManagerDaemon interface {
	// initiateControlSession iOS 11+
	initiateControlSession(XcodeVersion uint64) (err error)
	startExecutingTestPlan(XcodeVersion uint64) (err error)
	initiateSession(XcodeVersion uint64, nsUUID *nskeyedarchiver.NSUUID) (err error)
	// authorizeTestSession iOS 12+
	authorizeTestSession(pid int) (err error)
	// initiateControlSessionForTestProcessID <= iOS 9
	initiateControlSessionForTestProcessID(pid int) (err error)
	// initiateControlSessionForTestProcessIDProtocolVersion iOS > 9 && iOS < 12
	initiateControlSessionForTestProcessIDProtocolVersion(pid int, XcodeVersion uint64) (err error)

	registerCallback(obj string, cb func(m libimobiledevice.DTXMessageResult))
	close()
}

type SyslogRelay interface {
	Lines() <-chan string
	Stop()
}

type Pcapd interface {
	Packet() <-chan []byte
	Stop()
}

type Perfd interface {
	Start() (data <-chan []byte, err error)
	Stop()
}

type DiagnosticsRelay interface {
	Reboot() error
	Shutdown() error
}

type CrashReportMover interface {
	Move(hostDir string, opts ...CrashReportMoverOption) (err error)
	walkDir(dirname string, fn func(path string, info *AfcFileInfo)) (err error)
}

type SpringBoard interface {
	GetIconPNGData(bundleId string) (raw *bytes.Buffer, err error)
	GetInterfaceOrientation() (orientation OrientationState, err error)
}

type InnerConn = libimobiledevice.InnerConn

type LockdownType = libimobiledevice.LockdownType

type PairRecord = libimobiledevice.PairRecord

type CoordinateSystem = libimobiledevice.CoordinateSystem

const (
	CoordinateSystemWGS84 = libimobiledevice.CoordinateSystemWGS84
	CoordinateSystemBD09  = libimobiledevice.CoordinateSystemBD09
	CoordinateSystemGCJ02 = libimobiledevice.CoordinateSystemGCJ02
)

type ApplicationType = libimobiledevice.ApplicationType

const (
	ApplicationTypeSystem   = libimobiledevice.ApplicationTypeSystem
	ApplicationTypeUser     = libimobiledevice.ApplicationTypeUser
	ApplicationTypeInternal = libimobiledevice.ApplicationTypeInternal
	ApplicationTypeAny      = libimobiledevice.ApplicationTypeAny
)

type installationProxyOption = libimobiledevice.InstallationProxyOption

type InstallationProxyOption func(*installationProxyOption)

func WithApplicationType(appType ApplicationType) InstallationProxyOption {
	return func(opt *installationProxyOption) {
		opt.ApplicationType = appType
	}
}

func WithReturnAttributes(attrs ...string) InstallationProxyOption {
	return func(opt *installationProxyOption) {
		if len(opt.ReturnAttributes) == 0 {
			opt.ReturnAttributes = attrs
		} else {
			opt.ReturnAttributes = append(opt.ReturnAttributes, attrs...)
		}
		opt.ReturnAttributes = _removeDuplicate(opt.ReturnAttributes)
	}
}

func WithBundleIDs(BundleIDs ...string) InstallationProxyOption {
	return func(opt *installationProxyOption) {
		if len(opt.BundleIDs) == 0 {
			opt.BundleIDs = BundleIDs
		} else {
			opt.BundleIDs = append(opt.BundleIDs, BundleIDs...)
		}
		opt.BundleIDs = _removeDuplicate(opt.BundleIDs)
	}
}

func WithMetaData(b bool) InstallationProxyOption {
	return func(opt *installationProxyOption) {
		opt.MetaData = b
	}
}

type appLaunchOption struct {
	appPath     string
	environment map[string]interface{}
	arguments   []interface{}
	options     map[string]interface{}
}

type AppLaunchOption func(option *appLaunchOption)

func WithAppPath(appPath string) AppLaunchOption {
	return func(opt *appLaunchOption) {
		opt.appPath = appPath
	}
}

func WithEnvironment(environment map[string]interface{}) AppLaunchOption {
	return func(opt *appLaunchOption) {
		opt.environment = environment
	}
}

func WithArguments(arguments []interface{}) AppLaunchOption {
	return func(opt *appLaunchOption) {
		opt.arguments = arguments
	}
}

func WithOptions(options map[string]interface{}) AppLaunchOption {
	return func(opt *appLaunchOption) {
		for k, v := range options {
			opt.options[k] = v
		}
	}
}

func WithKillExisting(b bool) AppLaunchOption {
	return func(opt *appLaunchOption) {
		v := uint64(0)
		if b {
			v = uint64(1)
		}
		opt.options["KillExisting"] = v
	}
}

type appListOption struct {
	appsMatching map[string]interface{}
	updateToken  string
}

type AppListOption func(option *appListOption)

func WithAppsMatching(appsMatching map[string]interface{}) AppListOption {
	return func(opt *appListOption) {
		opt.appsMatching = appsMatching
	}
}

func WithUpdateToken(updateToken string) AppListOption {
	return func(opt *appListOption) {
		opt.updateToken = updateToken
	}
}

type Process struct {
	IsApplication bool      `json:"isApplication"`
	Name          string    `json:"name"`
	Pid           int       `json:"pid"`
	RealAppName   string    `json:"realAppName"`
	StartDate     time.Time `json:"startDate"`
}

type crashReportMoverOption struct {
	whenDone func(filename string)
	keep     bool
	extract  bool
}

func defaultCrashReportMoverOption() *crashReportMoverOption {
	return &crashReportMoverOption{
		whenDone: func(filename string) {},
		keep:     false,
	}
}

type CrashReportMoverOption func(opt *crashReportMoverOption)

func WithKeepCrashReport(b bool) CrashReportMoverOption {
	return func(opt *crashReportMoverOption) {
		opt.keep = b
	}
}

func WithExtractRawCrashReport(b bool) CrashReportMoverOption {
	return func(opt *crashReportMoverOption) {
		opt.extract = b
	}
}

func WithWhenMoveIsDone(whenDone func(filename string)) CrashReportMoverOption {
	return func(opt *crashReportMoverOption) {
		opt.whenDone = whenDone
	}
}

type xcTestOption struct {
	appEnv  map[string]interface{}
	appArgs []interface{}
	appOpt  map[string]interface{}
}

func defaultXCTestOption() *xcTestOption {
	return &xcTestOption{
		appEnv:  make(map[string]interface{}),
		appArgs: make([]interface{}, 0, 2),
		appOpt:  make(map[string]interface{}),
	}
}

type XCTestOption func(opt *xcTestOption)

func WithXCTestEnv(env map[string]interface{}) XCTestOption {
	return func(opt *xcTestOption) {
		opt.appEnv = env
	}
}

// func WithXCTestArgs(args []interface{}) XCTestOption {
// 	return func(opt *xcTestOption) {
// 		opt.appArgs = args
// 	}
// }

func WithXCTestOpt(appOpt map[string]interface{}) XCTestOption {
	return func(opt *xcTestOption) {
		opt.appOpt = appOpt
	}
}

func _removeDuplicate(strSlice []string) []string {
	existed := make(map[string]bool, len(strSlice))
	noRepeat := make([]string, 0, len(strSlice))
	for _, str := range strSlice {
		if _, ok := existed[str]; ok {
			continue
		}
		existed[str] = true
		noRepeat = append(noRepeat, str)
	}
	return noRepeat
}

func DeviceVersion(version ...int) int {
	if len(version) < 3 {
		tmp := make([]int, 3)
		copy(tmp, version)
		version = tmp
	}
	maj, min, patch := version[0], version[1], version[2]
	return ((maj & 0xFF) << 16) | ((min & 0xFF) << 8) | (patch & 0xFF)
}

var debugFlag = false

// SetDebug sets debug mode
func SetDebug(debug bool, libDebug ...bool) {
	debugFlag = debug
	if len(libDebug) >= 1 {
		libimobiledevice.SetDebug(libDebug[0])
	}
}

func debugLog(msg string) {
	if !debugFlag {
		return
	}
	fmt.Printf("[go-iDevice-debug] %s\n", msg)
}
