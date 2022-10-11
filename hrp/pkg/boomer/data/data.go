/*
 * Copyright 2020 gRPC authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package data

import (
	"embed"
	"os"
	"path/filepath"

	"github.com/httprunner/httprunner/v4/hrp/internal/builtin"
)

// hrpPath is .hrp directory under the user directory.
var hrpPath string

//go:embed x509/*
var x509Dir embed.FS

func init() {
	home, err := os.UserHomeDir()
	if err != nil {
		return
	}
	hrpPath = filepath.Join(home, ".hrp")
	_ = builtin.EnsureFolderExists(filepath.Join(hrpPath, "x509"))
}

// Path returns the absolute path the given relative file or directory path
func Path(rel string) (destPath string) {
	destPath = rel
	if !filepath.IsAbs(rel) {
		destPath = filepath.Join(hrpPath, rel)
	}
	if !builtin.IsFilePathExists(destPath) {
		content, err := x509Dir.ReadFile(rel)
		if err != nil {
			return
		}

		err = os.WriteFile(destPath, content, 0o644)
		if err != nil {
			return
		}
	}
	return
}
