package ghdc

import (
	"errors"
	"fmt"
	"net"
	"os"
	"strings"
	"time"
)

type DeviceFileInfo struct {
	Name         string
	Mode         os.FileMode
	Size         uint32
	LastModified time.Time
}

func (info DeviceFileInfo) IsDir() bool {
	return (info.Mode & (1 << 14)) == (1 << 14)
}

type DeviceState string

const (
	StateOnline  DeviceState = "Online"
	StateOffline DeviceState = "Offline"
)

var deviceStateStrings = map[string]DeviceState{
	"Offline":   StateOffline,
	"Connected": StateOnline,
}

type KeyCode int

const (
	KEYCODE_FN                       KeyCode = 0
	KEYCODE_UNKNOWN                  KeyCode = -1
	KEYCODE_HOME                     KeyCode = 1
	KEYCODE_BACK                     KeyCode = 2
	KEYCODE_MEDIA_PLAY_PAUSE         KeyCode = 10
	KEYCODE_MEDIA_STOP               KeyCode = 11
	KEYCODE_MEDIA_NEXT               KeyCode = 12
	KEYCODE_MEDIA_PREVIOUS           KeyCode = 13
	KEYCODE_MEDIA_REWIND             KeyCode = 14
	KEYCODE_MEDIA_FAST_FORWARD       KeyCode = 15
	KEYCODE_VOLUME_UP                KeyCode = 16
	KEYCODE_VOLUME_DOWN              KeyCode = 17
	KEYCODE_POWER                    KeyCode = 18
	KEYCODE_CAMERA                   KeyCode = 19
	KEYCODE_VOLUME_MUTE              KeyCode = 22
	KEYCODE_MUTE                     KeyCode = 23
	KEYCODE_BRIGHTNESS_UP            KeyCode = 40
	KEYCODE_BRIGHTNESS_DOWN          KeyCode = 41
	KEYCODE_NUM_0                    KeyCode = 2000
	KEYCODE_NUM_1                    KeyCode = 2001
	KEYCODE_NUM_2                    KeyCode = 2002
	KEYCODE_NUM_3                    KeyCode = 2003
	KEYCODE_NUM_4                    KeyCode = 2004
	KEYCODE_NUM_5                    KeyCode = 2005
	KEYCODE_NUM_6                    KeyCode = 2006
	KEYCODE_NUM_7                    KeyCode = 2007
	KEYCODE_NUM_8                    KeyCode = 2008
	KEYCODE_NUM_9                    KeyCode = 2009
	KEYCODE_STAR                     KeyCode = 2010
	KEYCODE_POUND                    KeyCode = 2011
	KEYCODE_DPAD_UP                  KeyCode = 2012
	KEYCODE_DPAD_DOWN                KeyCode = 2013
	KEYCODE_DPAD_LEFT                KeyCode = 2014
	KEYCODE_DPAD_RIGHT               KeyCode = 2015
	KEYCODE_DPAD_CENTER              KeyCode = 2016
	KEYCODE_A                        KeyCode = 2017
	KEYCODE_B                        KeyCode = 2018
	KEYCODE_C                        KeyCode = 2019
	KEYCODE_D                        KeyCode = 2020
	KEYCODE_E                        KeyCode = 2021
	KEYCODE_F                        KeyCode = 2022
	KEYCODE_G                        KeyCode = 2023
	KEYCODE_H                        KeyCode = 2024
	KEYCODE_I                        KeyCode = 2025
	KEYCODE_J                        KeyCode = 2026
	KEYCODE_K                        KeyCode = 2027
	KEYCODE_L                        KeyCode = 2028
	KEYCODE_M                        KeyCode = 2029
	KEYCODE_N                        KeyCode = 2030
	KEYCODE_O                        KeyCode = 2031
	KEYCODE_P                        KeyCode = 2032
	KEYCODE_Q                        KeyCode = 2033
	KEYCODE_R                        KeyCode = 2034
	KEYCODE_S                        KeyCode = 2035
	KEYCODE_T                        KeyCode = 2036
	KEYCODE_U                        KeyCode = 2037
	KEYCODE_V                        KeyCode = 2038
	KEYCODE_W                        KeyCode = 2039
	KEYCODE_X                        KeyCode = 2040
	KEYCODE_Y                        KeyCode = 2041
	KEYCODE_Z                        KeyCode = 2042
	KEYCODE_COMMA                    KeyCode = 2043
	KEYCODE_PERIOD                   KeyCode = 2044
	KEYCODE_ALT_LEFT                 KeyCode = 2045
	KEYCODE_ALT_RIGHT                KeyCode = 2046
	KEYCODE_SHIFT_LEFT               KeyCode = 2047
	KEYCODE_SHIFT_RIGHT              KeyCode = 2048
	KEYCODE_TAB                      KeyCode = 2049
	KEYCODE_SPACE                    KeyCode = 2050
	KEYCODE_SYM                      KeyCode = 2051
	KEYCODE_EXPLORER                 KeyCode = 2052
	KEYCODE_ENVELOPE                 KeyCode = 2053
	KEYCODE_ENTER                    KeyCode = 2054
	KEYCODE_DEL                      KeyCode = 2055
	KEYCODE_GRAVE                    KeyCode = 2056
	KEYCODE_MINUS                    KeyCode = 2057
	KEYCODE_EQUALS                   KeyCode = 2058
	KEYCODE_LEFT_BRACKET             KeyCode = 2059
	KEYCODE_RIGHT_BRACKET            KeyCode = 2060
	KEYCODE_BACKSLASH                KeyCode = 2061
	KEYCODE_SEMICOLON                KeyCode = 2062
	KEYCODE_APOSTROPHE               KeyCode = 2063
	KEYCODE_SLASH                    KeyCode = 2064
	KEYCODE_AT                       KeyCode = 2065
	KEYCODE_PLUS                     KeyCode = 2066
	KEYCODE_MENU                     KeyCode = 2067
	KEYCODE_PAGE_UP                  KeyCode = 2068
	KEYCODE_PAGE_DOWN                KeyCode = 2069
	KEYCODE_ESCAPE                   KeyCode = 2070
	KEYCODE_FORWARD_DEL              KeyCode = 2071
	KEYCODE_CTRL_LEFT                KeyCode = 2072
	KEYCODE_CTRL_RIGHT               KeyCode = 2073
	KEYCODE_CAPS_LOCK                KeyCode = 2074
	KEYCODE_SCROLL_LOCK              KeyCode = 2075
	KEYCODE_META_LEFT                KeyCode = 2076
	KEYCODE_META_RIGHT               KeyCode = 2077
	KEYCODE_FUNCTION                 KeyCode = 2078
	KEYCODE_SYSRQ                    KeyCode = 2079
	KEYCODE_BREAK                    KeyCode = 2080
	KEYCODE_MOVE_HOME                KeyCode = 2081
	KEYCODE_MOVE_END                 KeyCode = 2082
	KEYCODE_INSERT                   KeyCode = 2083
	KEYCODE_FORWARD                  KeyCode = 2084
	KEYCODE_MEDIA_PLAY               KeyCode = 2085
	KEYCODE_MEDIA_PAUSE              KeyCode = 2086
	KEYCODE_MEDIA_CLOSE              KeyCode = 2087
	KEYCODE_MEDIA_EJECT              KeyCode = 2088
	KEYCODE_MEDIA_RECORD             KeyCode = 2089
	KEYCODE_F1                       KeyCode = 2090
	KEYCODE_F2                       KeyCode = 2091
	KEYCODE_F3                       KeyCode = 2092
	KEYCODE_F4                       KeyCode = 2093
	KEYCODE_F5                       KeyCode = 2094
	KEYCODE_F6                       KeyCode = 2095
	KEYCODE_F7                       KeyCode = 2096
	KEYCODE_F8                       KeyCode = 2097
	KEYCODE_F9                       KeyCode = 2098
	KEYCODE_F10                      KeyCode = 2099
	KEYCODE_F11                      KeyCode = 2100
	KEYCODE_F12                      KeyCode = 2101
	KEYCODE_NUM_LOCK                 KeyCode = 2102
	KEYCODE_NUMPAD_0                 KeyCode = 2103
	KEYCODE_NUMPAD_1                 KeyCode = 2104
	KEYCODE_NUMPAD_2                 KeyCode = 2105
	KEYCODE_NUMPAD_3                 KeyCode = 2106
	KEYCODE_NUMPAD_4                 KeyCode = 2107
	KEYCODE_NUMPAD_5                 KeyCode = 2108
	KEYCODE_NUMPAD_6                 KeyCode = 2109
	KEYCODE_NUMPAD_7                 KeyCode = 2110
	KEYCODE_NUMPAD_8                 KeyCode = 2111
	KEYCODE_NUMPAD_9                 KeyCode = 2112
	KEYCODE_NUMPAD_DIVIDE            KeyCode = 2113
	KEYCODE_NUMPAD_MULTIPLY          KeyCode = 2114
	KEYCODE_NUMPAD_SUBTRACT          KeyCode = 2115
	KEYCODE_NUMPAD_ADD               KeyCode = 2116
	KEYCODE_NUMPAD_DOT               KeyCode = 2117
	KEYCODE_NUMPAD_COMMA             KeyCode = 2118
	KEYCODE_NUMPAD_ENTER             KeyCode = 2119
	KEYCODE_NUMPAD_EQUALS            KeyCode = 2120
	KEYCODE_NUMPAD_LEFT_PAREN        KeyCode = 2121
	KEYCODE_NUMPAD_RIGHT_PAREN       KeyCode = 2122
	KEYCODE_VIRTUAL_MULTITASK        KeyCode = 2210
	KEYCODE_SLEEP                    KeyCode = 2600
	KEYCODE_ZENKAKU_HANKAKU          KeyCode = 2601
	KEYCODE_ND                       KeyCode = 2602
	KEYCODE_RO                       KeyCode = 2603
	KEYCODE_KATAKANA                 KeyCode = 2604
	KEYCODE_HIRAGANA                 KeyCode = 2605
	KEYCODE_HENKAN                   KeyCode = 2606
	KEYCODE_KATAKANA_HIRAGANA        KeyCode = 2607
	KEYCODE_MUHENKAN                 KeyCode = 2608
	KEYCODE_LINEFEED                 KeyCode = 2609
	KEYCODE_MACRO                    KeyCode = 2610
	KEYCODE_NUMPAD_PLUSMINUS         KeyCode = 2611
	KEYCODE_SCALE                    KeyCode = 2612
	KEYCODE_HANGUEL                  KeyCode = 2613
	KEYCODE_HANJA                    KeyCode = 2614
	KEYCODE_YEN                      KeyCode = 2615
	KEYCODE_STOP                     KeyCode = 2616
	KEYCODE_AGAIN                    KeyCode = 2617
	KEYCODE_PROPS                    KeyCode = 2618
	KEYCODE_UNDO                     KeyCode = 2619
	KEYCODE_COPY                     KeyCode = 2620
	KEYCODE_OPEN                     KeyCode = 2621
	KEYCODE_PASTE                    KeyCode = 2622
	KEYCODE_FIND                     KeyCode = 2623
	KEYCODE_CUT                      KeyCode = 2624
	KEYCODE_HELP                     KeyCode = 2625
	KEYCODE_CALC                     KeyCode = 2626
	KEYCODE_FILE                     KeyCode = 2627
	KEYCODE_BOOKMARKS                KeyCode = 2628
	KEYCODE_NEXT                     KeyCode = 2629
	KEYCODE_PLAYPAUSE                KeyCode = 2630
	KEYCODE_PREVIOUS                 KeyCode = 2631
	KEYCODE_STOPCD                   KeyCode = 2632
	KEYCODE_CONFIG                   KeyCode = 2634
	KEYCODE_REFRESH                  KeyCode = 2635
	KEYCODE_EXIT                     KeyCode = 2636
	KEYCODE_EDIT                     KeyCode = 2637
	KEYCODE_SCROLLUP                 KeyCode = 2638
	KEYCODE_SCROLLDOWN               KeyCode = 2639
	KEYCODE_NEW                      KeyCode = 2640
	KEYCODE_REDO                     KeyCode = 2641
	KEYCODE_CLOSE                    KeyCode = 2642
	KEYCODE_PLAY                     KeyCode = 2643
	KEYCODE_BASSBOOST                KeyCode = 2644
	KEYCODE_PRINT                    KeyCode = 2645
	KEYCODE_CHAT                     KeyCode = 2646
	KEYCODE_FINANCE                  KeyCode = 2647
	KEYCODE_CANCEL                   KeyCode = 2648
	KEYCODE_KBDILLUM_TOGGLE          KeyCode = 2649
	KEYCODE_KBDILLUM_DOWN            KeyCode = 2650
	KEYCODE_KBDILLUM_UP              KeyCode = 2651
	KEYCODE_SEND                     KeyCode = 2652
	KEYCODE_REPLY                    KeyCode = 2653
	KEYCODE_FORWARDMAIL              KeyCode = 2654
	KEYCODE_SAVE                     KeyCode = 2655
	KEYCODE_DOCUMENTS                KeyCode = 2656
	KEYCODE_VIDEO_NEXT               KeyCode = 2657
	KEYCODE_VIDEO_PREV               KeyCode = 2658
	KEYCODE_BRIGHTNESS_CYCLE         KeyCode = 2659
	KEYCODE_BRIGHTNESS_ZERO          KeyCode = 2660
	KEYCODE_DISPLAY_OFF              KeyCode = 2661
	KEYCODE_BTN_MISC                 KeyCode = 2662
	KEYCODE_GOTO                     KeyCode = 2663
	KEYCODE_INFO                     KeyCode = 2664
	KEYCODE_PROGRAM                  KeyCode = 2665
	KEYCODE_PVR                      KeyCode = 2666
	KEYCODE_SUBTITLE                 KeyCode = 2667
	KEYCODE_FULL_SCREEN              KeyCode = 2668
	KEYCODE_KEYBOARD                 KeyCode = 2669
	KEYCODE_ASPECT_RATIO             KeyCode = 2670
	KEYCODE_PC                       KeyCode = 2671
	KEYCODE_TV                       KeyCode = 2672
	KEYCODE_TV2                      KeyCode = 2673
	KEYCODE_VCR                      KeyCode = 2674
	KEYCODE_VCR2                     KeyCode = 2675
	KEYCODE_SAT                      KeyCode = 2676
	KEYCODE_CD                       KeyCode = 2677
	KEYCODE_TAPE                     KeyCode = 2678
	KEYCODE_TUNER                    KeyCode = 2679
	KEYCODE_PLAYER                   KeyCode = 2680
	KEYCODE_DVD                      KeyCode = 2681
	KEYCODE_AUDIO                    KeyCode = 2682
	KEYCODE_VIDEO                    KeyCode = 2683
	KEYCODE_MEMO                     KeyCode = 2684
	KEYCODE_CALENDAR                 KeyCode = 2685
	KEYCODE_RED                      KeyCode = 2686
	KEYCODE_GREEN                    KeyCode = 2687
	KEYCODE_YELLOW                   KeyCode = 2688
	KEYCODE_BLUE                     KeyCode = 2689
	KEYCODE_CHANNELUP                KeyCode = 2690
	KEYCODE_CHANNELDOWN              KeyCode = 2691
	KEYCODE_LAST                     KeyCode = 2692
	KEYCODE_RESTART                  KeyCode = 2693
	KEYCODE_SLOW                     KeyCode = 2694
	KEYCODE_SHUFFLE                  KeyCode = 2695
	KEYCODE_VIDEOPHONE               KeyCode = 2696
	KEYCODE_GAMES                    KeyCode = 2697
	KEYCODE_ZOOMIN                   KeyCode = 2698
	KEYCODE_ZOOMOUT                  KeyCode = 2699
	KEYCODE_ZOOMRESET                KeyCode = 2700
	KEYCODE_WORDPROCESSOR            KeyCode = 2701
	KEYCODE_EDITOR                   KeyCode = 2702
	KEYCODE_SPREADSHEET              KeyCode = 2703
	KEYCODE_GRAPHICSEDITOR           KeyCode = 2704
	KEYCODE_PRESENTATION             KeyCode = 2705
	KEYCODE_DATABASE                 KeyCode = 2706
	KEYCODE_NEWS                     KeyCode = 2707
	KEYCODE_VOICEMAIL                KeyCode = 2708
	KEYCODE_ADDRESSBOOK              KeyCode = 2709
	KEYCODE_MESSENGER                KeyCode = 2710
	KEYCODE_BRIGHTNESS_TOGGLE        KeyCode = 2711
	KEYCODE_SPELLCHECK               KeyCode = 2712
	KEYCODE_COFFEE                   KeyCode = 2713
	KEYCODE_MEDIA_REPEAT             KeyCode = 2714
	KEYCODE_IMAGES                   KeyCode = 2715
	KEYCODE_BUTTONCONFIG             KeyCode = 2716
	KEYCODE_TASKMANAGER              KeyCode = 2717
	KEYCODE_JOURNAL                  KeyCode = 2718
	KEYCODE_CONTROLPANEL             KeyCode = 2719
	KEYCODE_APPSELECT                KeyCode = 2720
	KEYCODE_SCREENSAVER              KeyCode = 2721
	KEYCODE_ASSISTANT                KeyCode = 2722
	KEYCODE_KBD_LAYOUT_NEXT          KeyCode = 2723
	KEYCODE_BRIGHTNESS_MIN           KeyCode = 2724
	KEYCODE_BRIGHTNESS_MAX           KeyCode = 2725
	KEYCODE_KBDINPUTASSIST_PREV      KeyCode = 2726
	KEYCODE_KBDINPUTASSIST_NEXT      KeyCode = 2727
	KEYCODE_KBDINPUTASSIST_PREVGROUP KeyCode = 2728
	KEYCODE_KBDINPUTASSIST_NEXTGROUP KeyCode = 2729
	KEYCODE_KBDINPUTASSIST_ACCEPT    KeyCode = 2730
	KEYCODE_KBDINPUTASSIST_CANCEL    KeyCode = 2731
	KEYCODE_FRONT                    KeyCode = 2800
	KEYCODE_SETUP                    KeyCode = 2801
	KEYCODE_WAKE_UP                  KeyCode = 2802
	KEYCODE_SENDFILE                 KeyCode = 2803
	KEYCODE_DELETEFILE               KeyCode = 2804
	KEYCODE_XFER                     KeyCode = 2805
	KEYCODE_PROG1                    KeyCode = 2806
	KEYCODE_PROG2                    KeyCode = 2807
	KEYCODE_MSDOS                    KeyCode = 2808
	KEYCODE_SCREENLOCK               KeyCode = 2809
	KEYCODE_DIRECTION_ROTATE_DISPLAY KeyCode = 2810
	KEYCODE_CYCLEWINDOWS             KeyCode = 2811
	KEYCODE_COMPUTER                 KeyCode = 2812
	KEYCODE_EJECTCLOSECD             KeyCode = 2813
	KEYCODE_ISO                      KeyCode = 2814
	KEYCODE_MOVE                     KeyCode = 2815
	KEYCODE_F13                      KeyCode = 2816
	KEYCODE_F14                      KeyCode = 2817
	KEYCODE_F15                      KeyCode = 2818
	KEYCODE_F16                      KeyCode = 2819
	KEYCODE_F17                      KeyCode = 2820
	KEYCODE_F18                      KeyCode = 2821
	KEYCODE_F19                      KeyCode = 2822
	KEYCODE_F20                      KeyCode = 2823
	KEYCODE_F21                      KeyCode = 2824
	KEYCODE_F22                      KeyCode = 2825
	KEYCODE_F23                      KeyCode = 2826
	KEYCODE_F24                      KeyCode = 2827
	KEYCODE_PROG3                    KeyCode = 2828
	KEYCODE_PROG4                    KeyCode = 2829
	KEYCODE_DASHBOARD                KeyCode = 2830
	KEYCODE_SUSPEND                  KeyCode = 2831
	KEYCODE_HP                       KeyCode = 2832
	KEYCODE_SOUND                    KeyCode = 2833
	KEYCODE_QUESTION                 KeyCode = 2834
	KEYCODE_CONNECT                  KeyCode = 2836
	KEYCODE_SPORT                    KeyCode = 2837
	KEYCODE_SHOP                     KeyCode = 2838
	KEYCODE_ALTERASE                 KeyCode = 2839
	KEYCODE_SWITCHVIDEOMODE          KeyCode = 2841
	KEYCODE_BATTERY                  KeyCode = 2842
	KEYCODE_BLUETOOTH                KeyCode = 2843
	KEYCODE_WLAN                     KeyCode = 2844
	KEYCODE_UWB                      KeyCode = 2845
	KEYCODE_WWAN_WIMAX               KeyCode = 2846
	KEYCODE_RFKILL                   KeyCode = 2847
	KEYCODE_CHANNEL                  KeyCode = 3001
	KEYCODE_BTN_0                    KeyCode = 3100
	KEYCODE_BTN_1                    KeyCode = 3101
	KEYCODE_BTN_2                    KeyCode = 3102
	KEYCODE_BTN_3                    KeyCode = 3103
	KEYCODE_BTN_4                    KeyCode = 3104
	KEYCODE_BTN_5                    KeyCode = 3105
	KEYCODE_BTN_6                    KeyCode = 3106
	KEYCODE_BTN_7                    KeyCode = 3107
	KEYCODE_BTN_8                    KeyCode = 3108
	KEYCODE_BTN_9                    KeyCode = 3109
)

type DeviceForward struct {
	Local  string
	Remote string
}

type Device struct {
	hdClient Client
	serial   string
	attrs    map[string]string
}

func NewDevice(hdClient Client, serial string, attrs map[string]string) (Device, error) {
	device := Device{hdClient: hdClient, serial: serial, attrs: attrs}
	model, err := device.RunShellCommand("param get const.product.model")
	if err != nil {
		return device, err
	}
	attrs["model"] = model

	brand, err := device.RunShellCommand("param get const.product.brand")
	if err != nil {
		return device, err
	}
	attrs["brand"] = brand

	sdkVersion, err := device.RunShellCommand("param get const.product.software.version")
	if err != nil {
		return device, err
	}
	attrs["sdkVersion"] = sdkVersion

	osVersion, err := device.RunShellCommand("param get const.ohos.apiversion")
	if err != nil {
		return device, err
	}
	attrs["osVersion"] = osVersion

	cpu, err := device.RunShellCommand("param get const.product.cpu.abilist")
	if err != nil {
		return device, err
	}
	attrs["cpu"] = cpu

	product, err := device.RunShellCommand("param get const.product.name")
	if err != nil {
		return device, err
	}
	attrs["product"] = product

	_, err = device.RunShellCommand("setenforce 1")
	if err != nil {
		return device, err
	}

	return device, nil
}

func (d Device) HasAttribute(key string) bool {
	_, ok := d.attrs[key]
	return ok
}

func (d Device) Product() (string, error) {
	if d.HasAttribute("product") {
		return d.attrs["product"], nil
	}
	return "", errors.New("does not have attribute: product")
}

func (d Device) Model() (string, error) {
	if d.HasAttribute("model") {
		return d.attrs["model"], nil
	}
	return "", errors.New("does not have attribute: model")
}

func (d Device) Usb() (string, error) {
	if d.HasAttribute("usb") {
		return d.attrs["usb"], nil
	}
	return "", errors.New("does not have attribute: usb")
}

func (d Device) DeviceInfo() map[string]string {
	return d.attrs
}

func (d Device) Serial() string {
	return d.serial
}

func (d Device) IsUsb() (bool, error) {
	usb, err := d.Usb()
	if err != nil {
		return false, err
	}

	return usb != "", nil
}

func (d Device) Screenshot(localPath string) error {
	tmpPath := fmt.Sprintf("/data/local/tmp/hypium_tmp_shot_%d.jpeg", time.Now().Unix())
	_, err := d.RunShellCommand("snapshot_display", "-f", tmpPath)
	if err != nil {
		err = fmt.Errorf("failed to take screencap \n%v", err)
		return err
	}
	err = d.PullFile(tmpPath, localPath)
	if err != nil {
		return err
	}
	_, _ = d.RunShellCommand("rm", "-rf", "tmpPath")
	return nil
}

func (d Device) Install(localPath string) error {
	res, err := d.ExecuteCommand(fmt.Sprintf("install -r %s", localPath))
	if err != nil || !strings.Contains(res, "success") {
		err = fmt.Errorf("failed to install %s %v, Msg: %s", localPath, err, res)
		return err
	}
	return nil
}

func (d Device) Forward(remotePort int) (localPort int, err error) {
	remote := fmt.Sprintf("tcp:%d", remotePort)
	localPort, err = GetFreePort()
	if err != nil {
		err = fmt.Errorf("failed to get free port \n%v", err)
		return
	}

	command := ""
	local := fmt.Sprintf("tcp:%d", localPort)

	command = fmt.Sprintf("fport %s %s", local, remote)
	_, err = d.ExecuteCommand(command)
	return
}

func (d Device) ForwardKill(localPort int) (err error) {
	local := fmt.Sprintf("tcp:%d", localPort)
	_, err = d.hdClient.executeCommand(fmt.Sprintf("-t %s fport rm %s", d.serial, local))
	return
}

func (d Device) RunShellCommand(cmd string, args ...string) (string, error) {
	if len(args) > 0 {
		cmd = fmt.Sprintf("%s %s", cmd, strings.Join(args, " "))
	}
	if strings.TrimSpace(cmd) == "" {
		return "", errors.New("hd shell: command cannot be empty")
	}
	return d.ExecuteCommand(fmt.Sprintf("shell %s", cmd))
}

func (d Device) createDeviceTransport() (tp transport, err error) {
	return newTransport(fmt.Sprintf("%s:%d", d.hdClient.host, d.hdClient.port), false, d.serial)
}

func (d Device) createDeviceAliveTransport() (tp transport, err error) {
	return newTransport(fmt.Sprintf("%s:%d", d.hdClient.host, d.hdClient.port), true, d.serial)
}

func (d Device) createUitestTransport() (uTp uitestTransport, err error) {
	port, err := d.Forward(8012)
	if err != nil {
		err = fmt.Errorf("failed to forward uitest port \n%v", err)
		return
	}
	return newUitestTransport(d.hdClient.host, fmt.Sprintf("%d", port))
}

func (d Device) createUitestKitTransport() (uTp uitestKitTransport, err error) {
	port, err := d.Forward(8012)
	if err != nil {
		err = fmt.Errorf("failed to forward uitest port \n%v", err)
		return
	}
	return newUitestKitTransport(d.serial, d.hdClient.host, fmt.Sprintf("%d", port))
}

func (d Device) ExecuteCommand(command string) (resp string, err error) {
	var tp transport
	if tp, err = d.createDeviceTransport(); err != nil {
		return "", err
	}
	defer func() { _ = tp.Close() }()
	time.Sleep(1 * time.Millisecond)
	if err = tp.SendCommand(command); err != nil {
		return "", err
	}
	resp, err = tp.ReadStringAll()
	if err != nil {
		return
	}
	if strings.Contains(resp, "[Fail]") {
		return resp, fmt.Errorf("failed to execute command 「%s」 \nerror: %s", command, resp)
	}
	return resp, nil
}

func (d Device) PushFile(localPath string, remotePath string) (err error) {
	var tp transport
	if tp, err = d.createDeviceTransport(); err != nil {
		return err
	}
	defer func() { _ = tp.Close() }()
	if err = tp.SendCommand(fmt.Sprintf("file send %s %s", localPath, remotePath)); err != nil {
		return err
	}
	_, err = tp.ReadAll()
	return nil
}

func (d Device) PullFile(remotePath string, localPath string) (err error) {
	var tp transport
	if tp, err = d.createDeviceTransport(); err != nil {
		return err
	}
	defer func() { _ = tp.Close() }()

	if err = tp.SendCommand(fmt.Sprintf("file recv %s %s", remotePath, localPath)); err != nil {
		return err
	}
	res, err := tp.ReadStringAll()
	if err == nil {
		if strings.Contains(res, "Fail") {
			return fmt.Errorf("failed to pull: msg: %s", res)
		}
	}
	return nil
}

func GetFreePort() (int, error) {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		return 0, fmt.Errorf("resolve tcp addr failed \n%v", err)
	}

	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return 0, fmt.Errorf("listen tcp addr failed \n%v", err)
	}
	defer func() {
		_ = l.Close()
	}()
	return l.Addr().(*net.TCPAddr).Port, nil
}
