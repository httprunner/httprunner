package gidevice

import (
	"fmt"

	"github.com/httprunner/httprunner/v4/hrp/pkg/gidevice/pkg/libimobiledevice"
	"github.com/httprunner/httprunner/v4/hrp/pkg/gidevice/pkg/nskeyedarchiver"
)

var _ XCTestManagerDaemon = (*xcTestManagerDaemon)(nil)

func newXcTestManagerDaemon(testmanagerd Testmanagerd, channelCode uint32) *xcTestManagerDaemon {
	return &xcTestManagerDaemon{
		testmanagerd: testmanagerd,
		channelCode:  channelCode,
	}
}

type xcTestManagerDaemon struct {
	testmanagerd Testmanagerd
	channelCode  uint32
}

func (d *xcTestManagerDaemon) initiateControlSession(XcodeVersion uint64) (err error) {
	args := libimobiledevice.NewAuxBuffer()
	if err = args.AppendObject(XcodeVersion); err != nil {
		return err
	}

	selector := "_IDE_initiateControlSessionWithProtocolVersion:"

	var ret *libimobiledevice.DTXMessageResult
	if ret, err = d.testmanagerd.invoke(selector, args, d.channelCode, true); err != nil {
		return err
	}

	if nsErr, ok := ret.Obj.(libimobiledevice.NSError); ok {
		return fmt.Errorf("%s", nsErr.NSUserInfo.(map[string]interface{})["NSLocalizedDescription"])
	}
	return
}

func (d *xcTestManagerDaemon) startExecutingTestPlan(XcodeVersion uint64) (err error) {
	args := libimobiledevice.NewAuxBuffer()
	if err = args.AppendObject(XcodeVersion); err != nil {
		return err
	}

	selector := "_IDE_startExecutingTestPlanWithProtocolVersion:"

	if _, err = d.testmanagerd.invoke(selector, args, 0xFFFFFFFF, false); err != nil {
		return err
	}

	return
}

func (d *xcTestManagerDaemon) initiateSession(XcodeVersion uint64, nsUUID *nskeyedarchiver.NSUUID) (err error) {
	args := libimobiledevice.NewAuxBuffer()
	if err = args.AppendObject(nsUUID); err != nil {
		return err
	}
	if err = args.AppendObject(nsUUID.String() + "-Go-iDevice"); err != nil {
		return err
	}
	if err = args.AppendObject("/Applications/Xcode.app/Contents/Developer/usr/bin/xcodebuild"); err != nil {
		return err
	}
	if err = args.AppendObject(XcodeVersion); err != nil {
		return err
	}

	selector := "_IDE_initiateSessionWithIdentifier:forClient:atPath:protocolVersion:"

	var ret *libimobiledevice.DTXMessageResult
	if ret, err = d.testmanagerd.invoke(selector, args, d.channelCode, true); err != nil {
		return err
	}

	if nsErr, ok := ret.Obj.(libimobiledevice.NSError); ok {
		return fmt.Errorf("%s", nsErr.NSUserInfo.(map[string]interface{})["NSLocalizedDescription"])
	}

	return
}

func (d *xcTestManagerDaemon) authorizeTestSession(pid int) (err error) {
	args := libimobiledevice.NewAuxBuffer()
	if err = args.AppendObject(pid); err != nil {
		return err
	}

	selector := "_IDE_authorizeTestSessionWithProcessID:"

	var ret *libimobiledevice.DTXMessageResult
	if ret, err = d.testmanagerd.invoke(selector, args, d.channelCode, true); err != nil {
		return err
	}

	if nsErr, ok := ret.Obj.(libimobiledevice.NSError); ok {
		return fmt.Errorf("%s", nsErr.NSUserInfo.(map[string]interface{})["NSLocalizedDescription"])
	}
	return
}

func (d *xcTestManagerDaemon) initiateControlSessionForTestProcessID(pid int) (err error) {
	args := libimobiledevice.NewAuxBuffer()
	if err = args.AppendObject(pid); err != nil {
		return err
	}

	selector := "_IDE_initiateControlSessionForTestProcessID:"

	var ret *libimobiledevice.DTXMessageResult
	if ret, err = d.testmanagerd.invoke(selector, args, d.channelCode, true); err != nil {
		return err
	}

	if nsErr, ok := ret.Obj.(libimobiledevice.NSError); ok {
		return fmt.Errorf("%s", nsErr.NSUserInfo.(map[string]interface{})["NSLocalizedDescription"])
	}
	return
}

func (d *xcTestManagerDaemon) initiateControlSessionForTestProcessIDProtocolVersion(pid int, XcodeVersion uint64) (err error) {
	args := libimobiledevice.NewAuxBuffer()
	if err = args.AppendObject(pid); err != nil {
		return err
	}
	if err = args.AppendObject(XcodeVersion); err != nil {
		return err
	}

	selector := "_IDE_initiateControlSessionForTestProcessID:protocolVersion:"

	var ret *libimobiledevice.DTXMessageResult
	if ret, err = d.testmanagerd.invoke(selector, args, d.channelCode, true); err != nil {
		return err
	}

	if nsErr, ok := ret.Obj.(libimobiledevice.NSError); ok {
		return fmt.Errorf("%s", nsErr.NSUserInfo.(map[string]interface{})["NSLocalizedDescription"])
	}
	return
}

func (d *xcTestManagerDaemon) registerCallback(obj string, cb func(m libimobiledevice.DTXMessageResult)) {
	d.testmanagerd.registerCallback(obj, cb)
}

func (d *xcTestManagerDaemon) close() {
	d.testmanagerd.close()
}
