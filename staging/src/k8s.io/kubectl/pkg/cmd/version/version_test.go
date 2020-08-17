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

package version

import (
	"bytes"
	apimachineryversion "k8s.io/apimachinery/pkg/version"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"k8s.io/cli-runtime/pkg/genericclioptions"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
)

func TestNewCmdVersionWithoutConfigFile(t *testing.T) {
	tf := cmdutil.NewFactory(&genericclioptions.ConfigFlags{})
	streams, _, buf, _ := genericclioptions.NewTestIOStreams()
	o := NewOptions(streams)
	if err := o.Complete(tf, &cobra.Command{}); err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if err := o.Validate(); err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	// FIXME soltysh:
	// since we have defaulting to localhost:8080 in staging/src/k8s.io/client-go/tools/clientcmd/client_config.go#getDefaultServer
	// we need to ignore the localhost:8080 server, when above gets removed this should be dropped too
	if err := o.Run(); err != nil && !strings.Contains(err.Error(), "localhost:8080") {
		t.Errorf("Cannot execute version command: %v", err)
	}
	if !strings.Contains(buf.String(), "Client Version") {
		t.Errorf("unexpected output: %s", buf.String())
	}
}

func TestSkewWarning(t *testing.T) {
	o := &Options{}

	var testCases = []struct {
		description     string
		clientVersion   *apimachineryversion.Info
		serverVersion   *apimachineryversion.Info
		warningExpected bool
	}{
		{
			description:     "same version should not print warning",
			clientVersion:   &apimachineryversion.Info{Major: "1", Minor: "19"},
			serverVersion:   &apimachineryversion.Info{Major: "1", Minor: "19"},
			warningExpected: false,
		},
		{
			description:     "client one minor version behind server should not print warning",
			clientVersion:   &apimachineryversion.Info{Major: "1", Minor: "19"},
			serverVersion:   &apimachineryversion.Info{Major: "1", Minor: "20"},
			warningExpected: false,
		},
		{
			description:     "client one minor version ahead of server should not print warning",
			clientVersion:   &apimachineryversion.Info{Major: "1", Minor: "20"},
			serverVersion:   &apimachineryversion.Info{Major: "1", Minor: "19"},
			warningExpected: false,
		},
		{
			description:     "client two minor versions behind server should print warning",
			clientVersion:   &apimachineryversion.Info{Major: "1", Minor: "18"},
			serverVersion:   &apimachineryversion.Info{Major: "1", Minor: "20"},
			warningExpected: true,
		},
		{
			description:     "client two minor versions ahead of server should print warning",
			clientVersion:   &apimachineryversion.Info{Major: "1", Minor: "20"},
			serverVersion:   &apimachineryversion.Info{Major: "1", Minor: "18"},
			warningExpected: true,
		},
		{
			description:     "client one major version behind server should print warning",
			clientVersion:   &apimachineryversion.Info{Major: "1", Minor: "1"},
			serverVersion:   &apimachineryversion.Info{Major: "2", Minor: "1"},
			warningExpected: true,
		},
		{
			description:     "client one major version ahead of server should print warning",
			clientVersion:   &apimachineryversion.Info{Major: "2", Minor: "1"},
			serverVersion:   &apimachineryversion.Info{Major: "1", Minor: "1"},
			warningExpected: true,
		},
		{
			description:     "should handle no warning scenario when there are non-numeric characters in minor version",
			clientVersion:   &apimachineryversion.Info{Major: "1", Minor: "20+"},
			serverVersion:   &apimachineryversion.Info{Major: "1", Minor: "20foo"},
			warningExpected: false,
		},
		{
			description:     "should handle warning scenario when there are non-numeric characters in minor version",
			clientVersion:   &apimachineryversion.Info{Major: "1", Minor: "18+"},
			serverVersion:   &apimachineryversion.Info{Major: "1", Minor: "20foo"},
			warningExpected: true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			errOut := &bytes.Buffer{}
			o.ErrOut = errOut
			o.checkVersionSkew(tc.clientVersion, tc.serverVersion)
			if tc.warningExpected {
				if errOut.Len() == 0 {
					t.Fatalf("Warning was expected but not printed")
				}
			} else {
				if errOut.Len() > 0 {
					t.Fatalf("Warning was printed but not expected")
				}
			}
		})
	}
}
