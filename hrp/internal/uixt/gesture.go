package uixt

import (
	"image"
	"sort"

	"github.com/electricbubble/gwda"
)

func (dExt *DriverExt) GesturePassword(pathname string, password ...int) (err error) {
	var rects []image.Rectangle
	if rects, err = dExt.FindAllImageRect(pathname); err != nil {
		return err
	}

	sort.Slice(rects, func(i, j int) bool {
		if rects[i].Min.Y < rects[j].Min.Y {
			return true
		} else if rects[i].Min.Y == rects[j].Min.Y {
			if rects[i].Min.X < rects[j].Min.X {
				return true
			}
		}
		return false
	})

	touchActions := gwda.NewTouchActions(len(password)*2 + 1)
	for i := range password {
		x, y, width, height := dExt.MappingToRectInUIKit(rects[password[i]])
		x = x + width*0.5
		y = y + height*0.5

		if i == 0 {
			touchActions.Press(gwda.NewTouchActionPress().WithXYFloat(x, y)).
				Wait(0.2)
		} else {
			touchActions.MoveTo(gwda.NewTouchActionMoveTo().WithXYFloat(x, y)).
				Wait(0.2)
		}
	}
	touchActions.Release()

	return dExt.PerformTouchActions(touchActions)
}
