package ipa

import (
	"archive/zip"
	"fmt"
	"io"
	"path"

	"howett.net/plist"
)

func Info(ipaPath string) (info map[string]interface{}, err error) {
	var reader *zip.ReadCloser
	if reader, err = zip.OpenReader(ipaPath); err != nil {
		return nil, err
	}

	defer func() {
		err = reader.Close()
	}()

	for _, file := range reader.File {
		matched, _err := path.Match("Payload/*.app/Info.plist", file.Name)
		if _err != nil {
			err = _err
			continue
		}
		if !matched {
			continue
		}

		var rd io.ReadCloser
		if rd, _err = file.Open(); _err != nil {
			return nil, _err
		}
		data, _err := io.ReadAll(rd)
		if _err != nil {
			return nil, _err
		}

		info = make(map[string]interface{})
		_, _err = plist.Unmarshal(data, &info)
		if _err != nil {
			return nil, _err
		}
	}

	if err != nil && len(info) == 0 {
		return nil, fmt.Errorf("find Info.plist: %w", err)
	}

	return
}
