package uixt

import (
	"encoding/xml"
	"fmt"
	"regexp"
	"strconv"
)

type Attributes struct {
	Index         int     `xml:"index,attr"`
	Package       string  `xml:"package,attr"`
	Class         string  `xml:"class,attr"`
	Text          string  `xml:"text,attr"`
	ResourceId    string  `xml:"resource-id,attr"`
	Checkable     bool    `xml:"checkable,attr"`
	Checked       bool    `xml:"checked,attr"`
	Clickable     bool    `xml:"clickable,attr"`
	Enabled       bool    `xml:"enabled,attr"`
	Focusable     bool    `xml:"focusable,attr"`
	Focused       bool    `xml:"focused,attr"`
	LongClickable bool    `xml:"long-clickable,attr"`
	Password      bool    `xml:"password,attr"`
	Scrollable    bool    `xml:"scrollable,attr"`
	Selected      bool    `xml:"selected,attr"`
	Bounds        *Bounds `xml:"bounds,attr"`
	Displayed     bool    `xml:"displayed,attr"`
}

type Hierarchy struct {
	XMLName xml.Name `xml:"hierarchy"`
	Attributes
	Layout []Layout `xml:",any"`
}

type Layout struct {
	Attributes
	Layout []Layout `xml:",any"`
}

type Bounds struct {
	X1, Y1, X2, Y2 int
}

func (b *Bounds) Center() (float64, float64) {
	return float64(b.X1+b.X2) / 2, float64(b.Y1+b.Y2) / 2
}

func (b *Bounds) UnmarshalXMLAttr(attr xml.Attr) error {
	// 正则表达式用于解析格式为"[x1,y1][x2,y2]"
	re := regexp.MustCompile(`\[(\d+),(\d+)]\[(\d+),(\d+)]`)
	matches := re.FindStringSubmatch(attr.Value)
	if matches == nil {
		return fmt.Errorf("bounds format is incorrect")
	}
	// 转换字符串为整数
	b.X1, _ = strconv.Atoi(matches[1])
	b.Y1, _ = strconv.Atoi(matches[2])
	b.X2, _ = strconv.Atoi(matches[3])
	b.Y2, _ = strconv.Atoi(matches[4])
	return nil
}
