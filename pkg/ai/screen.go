package ai

func NewScreenShotOptions(opts ...ScreenShotOption) *ScreenShotOptions {
	options := &ScreenShotOptions{}
	for _, option := range opts {
		option(options)
	}
	return options
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

func (o *ScreenShotOptions) Options() []ScreenShotOption {
	options := make([]ScreenShotOption, 0)
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

type ScreenShotOption func(o *ScreenShotOptions)

func WithScreenShotOCR(ocrOn bool) ScreenShotOption {
	return func(o *ScreenShotOptions) {
		o.ScreenShotWithOCR = ocrOn
	}
}

func WithScreenShotUpload(uploadOn bool) ScreenShotOption {
	return func(o *ScreenShotOptions) {
		o.ScreenShotWithUpload = uploadOn
	}
}

func WithScreenShotLiveType(liveTypeOn bool) ScreenShotOption {
	return func(o *ScreenShotOptions) {
		o.ScreenShotWithLiveType = liveTypeOn
	}
}

func WithScreenShotLivePopularity(livePopularityOn bool) ScreenShotOption {
	return func(o *ScreenShotOptions) {
		o.ScreenShotWithLivePopularity = livePopularityOn
	}
}

func WithScreenShotUITypes(uiTypes ...string) ScreenShotOption {
	return func(o *ScreenShotOptions) {
		o.ScreenShotWithUITypes = uiTypes
	}
}

func WithScreenShotClosePopups(closeOn bool) ScreenShotOption {
	return func(o *ScreenShotOptions) {
		o.ScreenShotWithClosePopups = closeOn
	}
}

func WithScreenOCRCluster(ocrCluster string) ScreenShotOption {
	return func(o *ScreenShotOptions) {
		o.ScreenShotWithOCRCluster = ocrCluster
	}
}

func WithScreenShotFileName(fileName string) ScreenShotOption {
	return func(o *ScreenShotOptions) {
		o.ScreenShotFileName = fileName
	}
}

// (x1, y1) is the top left corner, (x2, y2) is the bottom right corner
// [x1, y1, x2, y2] in percentage of the screen
type Scope []float64

func (s Scope) ToAbs(windowSize Size) AbsScope {
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

func (s AbsScope) Option() ScreenFilterOption {
	return WithAbsScope(s[0], s[1], s[2], s[3])
}

func NewScreenFilterOptions(opts ...ScreenFilterOption) *ScreenFilterOptions {
	options := &ScreenFilterOptions{}
	for _, option := range opts {
		option(options)
	}
	return options
}

type ScreenFilterOptions struct {
	// scope related
	Scope    Scope    `json:"scope,omitempty" yaml:"scope,omitempty"`
	AbsScope AbsScope `json:"abs_scope,omitempty" yaml:"abs_scope,omitempty"`

	Regex             bool  `json:"regex,omitempty" yaml:"regex,omitempty"`                             // use regex to match text
	Offset            []int `json:"offset,omitempty" yaml:"offset,omitempty"`                           // used to tap offset of point
	OffsetRandomRange []int `json:"offset_random_range,omitempty" yaml:"offset_random_range,omitempty"` // set random range [min, max] for tap/swipe points
	Index             int   `json:"index,omitempty" yaml:"index,omitempty"`                             // index of the target element
	MatchOne          bool  `json:"match_one,omitempty" yaml:"match_one,omitempty"`                     // match one of the targets if existed
}

type ScreenFilterOption func(o *ScreenFilterOptions)

// WithScope inputs area of [(x1,y1), (x2,y2)]
// x1, y1, x2, y2 are all in [0, 1], which means the relative position of the screen
func WithScope(x1, y1, x2, y2 float64) ScreenFilterOption {
	return func(o *ScreenFilterOptions) {
		o.Scope = Scope{x1, y1, x2, y2}
	}
}

// WithAbsScope inputs area of [(x1,y1), (x2,y2)]
// x1, y1, x2, y2 are all absolute position of the screen
func WithAbsScope(x1, y1, x2, y2 int) ScreenFilterOption {
	return func(o *ScreenFilterOptions) {
		o.AbsScope = AbsScope{x1, y1, x2, y2}
	}
}

// tap [x, y] with offset [offsetX, offsetY]
func WithTapOffset(offsetX, offsetY int) ScreenFilterOption {
	return func(o *ScreenFilterOptions) {
		o.Offset = []int{offsetX, offsetY}
	}
}

func WithRegex(regex bool) ScreenFilterOption {
	return func(o *ScreenFilterOptions) {
		o.Regex = regex
	}
}

func WithMatchOne(matchOne bool) ScreenFilterOption {
	return func(o *ScreenFilterOptions) {
		o.MatchOne = matchOne
	}
}

func WithIndex(index int) ScreenFilterOption {
	return func(o *ScreenFilterOptions) {
		o.Index = index
	}
}
