package option

import "github.com/httprunner/httprunner/v5/uixt/types"

type ScreenOptions struct {
	ScreenShotOptions
	ScreenRecordOptions
	ScreenFilterOptions
}

type ScreenShotOptions struct {
	ScreenShotWithOCR            bool     `json:"screenshot_with_ocr,omitempty" yaml:"screenshot_with_ocr,omitempty"`
	ScreenShotWithUpload         bool     `json:"screenshot_with_upload,omitempty" yaml:"screenshot_with_upload,omitempty"`
	ScreenShotWithLiveType       bool     `json:"screenshot_with_live_type,omitempty" yaml:"screenshot_with_live_type,omitempty"`
	ScreenShotWithLivePopularity bool     `json:"screenshot_with_live_popularity,omitempty" yaml:"screenshot_with_live_popularity,omitempty"`
	ScreenShotWithUITypes        []string `json:"screenshot_with_ui_types,omitempty" yaml:"screenshot_with_ui_types,omitempty"`
	ScreenShotWithClosePopups    bool     `json:"screenshot_with_close_popups,omitempty" yaml:"screenshot_with_close_popups,omitempty"`
	ScreenShotWithOCRCluster     string   `json:"screenshot_with_ocr_cluster,omitempty" yaml:"screenshot_with_ocr_cluster,omitempty"`
	ScreenShotFileName           string   `json:"screenshot_file_name,omitempty" yaml:"screenshot_file_name,omitempty"`
}

func (o *ScreenShotOptions) GetScreenShotOptions() []ActionOption {
	options := make([]ActionOption, 0)
	if o == nil {
		return options
	}

	// screenshot options
	if o.ScreenShotWithOCR {
		options = append(options, WithScreenShotOCR(true))
	}
	if o.ScreenShotWithUpload {
		options = append(options, WithScreenShotUpload(true))
	}
	if o.ScreenShotWithLiveType {
		options = append(options, WithScreenShotLiveType(true))
	}
	if o.ScreenShotWithLivePopularity {
		options = append(options, WithScreenShotLivePopularity(true))
	}
	if len(o.ScreenShotWithUITypes) > 0 {
		options = append(options, WithScreenShotUITypes(o.ScreenShotWithUITypes...))
	}
	if o.ScreenShotWithClosePopups {
		options = append(options, WithScreenShotClosePopups(true))
	}
	if o.ScreenShotWithOCRCluster != "" {
		options = append(options, WithScreenOCRCluster(o.ScreenShotWithOCRCluster))
	}
	if o.ScreenShotFileName != "" {
		options = append(options, WithScreenShotFileName(o.ScreenShotFileName))
	}

	return options
}

func (o *ScreenShotOptions) List() []string {
	options := []string{}
	if o.ScreenShotWithUpload {
		options = append(options, "upload")
	}
	if o.ScreenShotWithOCR {
		options = append(options, "ocr")
	}
	if o.ScreenShotWithLiveType {
		options = append(options, "liveType")
	}
	if o.ScreenShotWithLivePopularity {
		options = append(options, "livePopularity")
	}
	// UI detection
	if len(o.ScreenShotWithUITypes) > 0 {
		options = append(options, "ui")
	}
	if o.ScreenShotWithClosePopups {
		options = append(options, "close")
	}
	return options
}

func WithScreenShotOCR(ocrOn bool) ActionOption {
	return func(o *ActionOptions) {
		o.ScreenShotWithOCR = ocrOn
	}
}

func WithScreenShotUpload(uploadOn bool) ActionOption {
	return func(o *ActionOptions) {
		o.ScreenShotWithUpload = uploadOn
	}
}

func WithScreenShotLiveType(liveTypeOn bool) ActionOption {
	return func(o *ActionOptions) {
		o.ScreenShotWithLiveType = liveTypeOn
	}
}

func WithScreenShotLivePopularity(livePopularityOn bool) ActionOption {
	return func(o *ActionOptions) {
		o.ScreenShotWithLivePopularity = livePopularityOn
	}
}

func WithScreenShotUITypes(uiTypes ...string) ActionOption {
	return func(o *ActionOptions) {
		o.ScreenShotWithUITypes = uiTypes
	}
}

func WithScreenShotClosePopups(closeOn bool) ActionOption {
	return func(o *ActionOptions) {
		o.ScreenShotWithClosePopups = closeOn
	}
}

func WithScreenOCRCluster(ocrCluster string) ActionOption {
	return func(o *ActionOptions) {
		o.ScreenShotWithOCRCluster = ocrCluster
	}
}

func WithScreenShotFileName(fileName string) ActionOption {
	return func(o *ActionOptions) {
		o.ScreenShotFileName = fileName
	}
}

type ScreenRecordOptions struct {
	ScreenRecordDuration  float64 `json:"screenrecord_duration,omitempty" yaml:"screenrecord_duration,omitempty"`
	ScreenRecordWithAudio bool    `json:"screenrecord_with_audio,omitempty" yaml:"screenrecord_with_audio,omitempty"`
	ScreenRecordPath      string  `json:"screenrecord_path,omitempty" yaml:"screenrecord_path,omitempty"`
}

func (o *ScreenRecordOptions) GetScreenRecordOptions() []ActionOption {
	options := make([]ActionOption, 0)
	if o == nil {
		return options
	}

	// screen record options
	if o.ScreenRecordDuration > 0 {
		options = append(options, WithDuration(o.ScreenRecordDuration))
	}
	if o.ScreenRecordWithAudio {
		options = append(options, WithScreenRecordAudio(true))
	}
	if o.ScreenRecordPath != "" {
		options = append(options, WithScreenRecordPath(o.ScreenRecordPath))
	}
	return options
}

func WithScreenRecordDuation(duration float64) ActionOption {
	return func(o *ActionOptions) {
		o.ScreenRecordDuration = duration
	}
}

func WithScreenRecordAudio(audioOn bool) ActionOption {
	return func(o *ActionOptions) {
		o.ScreenRecordWithAudio = audioOn
	}
}

func WithScreenRecordPath(path string) ActionOption {
	return func(o *ActionOptions) {
		o.ScreenRecordPath = path
	}
}

// (x1, y1) is the top left corner, (x2, y2) is the bottom right corner
// [x1, y1, x2, y2] in percentage of the screen
type Scope []float64

func (s Scope) ToAbs(windowSize types.Size) AbsScope {
	x1, y1, x2, y2 := s[0], s[1], s[2], s[3]
	// convert relative scope to absolute scope
	absX1 := int(x1 * float64(windowSize.Width))
	absY1 := int(y1 * float64(windowSize.Height))
	absX2 := int(x2 * float64(windowSize.Width))
	absY2 := int(y2 * float64(windowSize.Height))
	return AbsScope{absX1, absY1, absX2, absY2}
}

// [x1, y1, x2, y2] in absolute pixels
type AbsScope []int

func (s AbsScope) Option() ActionOption {
	return WithAbsScope(s[0], s[1], s[2], s[3])
}

func NewScreenFilterOptions(opts ...ActionOption) *ActionOptions {
	options := &ActionOptions{}
	for _, option := range opts {
		option(options)
	}
	return options
}

type ScreenFilterOptions struct {
	// scope related
	Scope    Scope    `json:"scope,omitempty" yaml:"scope,omitempty"`
	AbsScope AbsScope `json:"abs_scope,omitempty" yaml:"abs_scope,omitempty"`

	Regex               bool  `json:"regex,omitempty" yaml:"regex,omitempty"`                             // use regex to match text
	Offset              []int `json:"offset,omitempty" yaml:"offset,omitempty"`                           // used to tap offset of point
	OffsetRandomRange   []int `json:"offset_random_range,omitempty" yaml:"offset_random_range,omitempty"` // set random range [min, max] for tap/swipe points
	Index               int   `json:"index,omitempty" yaml:"index,omitempty"`                             // index of the target element
	MatchOne            bool  `json:"match_one,omitempty" yaml:"match_one,omitempty"`
	IgnoreNotFoundError bool  `json:"ignore_NotFoundError,omitempty" yaml:"ignore_NotFoundError,omitempty"` // ignore error if target element not found                  // match one of the targets if existed
}

// WithScope inputs area of [(x1,y1), (x2,y2)]
// x1, y1, x2, y2 are all in [0, 1], which means the relative position of the screen
func WithScope(x1, y1, x2, y2 float64) ActionOption {
	return func(o *ActionOptions) {
		o.Scope = Scope{x1, y1, x2, y2}
	}
}

// WithAbsScope inputs area of [(x1,y1), (x2,y2)]
// x1, y1, x2, y2 are all absolute position of the screen
func WithAbsScope(x1, y1, x2, y2 int) ActionOption {
	return func(o *ActionOptions) {
		o.AbsScope = AbsScope{x1, y1, x2, y2}
	}
}

// tap [x, y] with offset [offsetX, offsetY]
func WithTapOffset(offsetX, offsetY int) ActionOption {
	return func(o *ActionOptions) {
		o.Offset = []int{offsetX, offsetY}
	}
}

func WithRegex(regex bool) ActionOption {
	return func(o *ActionOptions) {
		o.Regex = regex
	}
}

func WithMatchOne(matchOne bool) ActionOption {
	return func(o *ActionOptions) {
		o.MatchOne = matchOne
	}
}

func WithIndex(index int) ActionOption {
	return func(o *ActionOptions) {
		o.Index = index
	}
}
