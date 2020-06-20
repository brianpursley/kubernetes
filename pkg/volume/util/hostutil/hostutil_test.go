/*
Copyright 2020 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package hostutil

import (
	"errors"
	"os"
	"testing"
)

func TestGetFileType_PreserveErrorTypeWhenFileNotFound(t *testing.T) {
	fileType, err := getFileType("/tmp/i-do-not-exist")

	if fileType != FileTypeUnknown {
		t.Errorf("Expected file type to be unknown, but it was %v", fileType)
	}
	if !errors.Is(err, os.ErrNotExist) {
		t.Error("Expected error to be (or wrap) os.ErrNotExist, but it was not")
	}
}
