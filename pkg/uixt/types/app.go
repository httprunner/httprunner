package types

type AppInfo struct {
	Name string `json:"name,omitempty"`
	AppBaseInfo
}

type WindowInfo struct {
	PackageName string `json:"packageName,omitempty"`
	Activity    string `json:"activity,omitempty"`
}

type AppBaseInfo struct {
	Pid            int         `json:"pid,omitempty"`
	BundleId       string      `json:"bundleId,omitempty"`       // ios package name
	ViewController string      `json:"viewController,omitempty"` // ios view controller
	PackageName    string      `json:"packageName,omitempty"`    // android package name
	Activity       string      `json:"activity,omitempty"`       // android activity
	VersionName    string      `json:"versionName,omitempty"`
	VersionCode    interface{} `json:"versionCode,omitempty"` // int or string
	AppName        string      `json:"appName,omitempty"`
	AppPath        string      `json:"appPath,omitempty"`
	AppMD5         string      `json:"appMD5,omitempty"`
	// AppIcon        string `json:"appIcon,omitempty"`
}

type AppState int

const (
	AppStateNotRunning AppState = 1 << iota
	AppStateRunningBack
	AppStateRunningFront
)

func (v AppState) String() string {
	switch v {
	case AppStateNotRunning:
		return "Not Running"
	case AppStateRunningBack:
		return "Running (Back)"
	case AppStateRunningFront:
		return "Running (Front)"
	default:
		return "UNKNOWN"
	}
}

// PasteboardType The type of the item on the pasteboard.
type PasteboardType string

const (
	PasteboardTypePlaintext PasteboardType = "plaintext"
	PasteboardTypeImage     PasteboardType = "image"
	PasteboardTypeUrl       PasteboardType = "url"
)

const (
	TextBackspace string = "\u0008"
	TextDelete    string = "\u007F"
)
