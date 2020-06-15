package completion

import (
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/resource"
	"k8s.io/client-go/rest/fake"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	cmdtesting "k8s.io/kubectl/pkg/cmd/testing"
	"k8s.io/kubectl/pkg/scheme"
)

type completionFunction func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective)

func TestRegisterDynamicCompletions(t *testing.T) {
	flagsToCheck := []string{"cluster", "context", "namespace", "user"}
	commandsToCheck := []string{"get", "describe", "logs", "attach"}

	cmd := &cobra.Command{}
	for _, flagName := range flagsToCheck {
		cmd.Flags().String(flagName, "", "")
	}

	tf := cmdtesting.NewTestFactory().WithNamespace("test")
	pathOptions := clientcmd.NewDefaultPathOptions()

	RegisterDynamicCompletions(cmd, tf, pathOptions)

	// Check some commands to make sure they have a valid args function
	for _, commandText := range commandsToCheck {
		subCommand, _, _ := cmd.Find(strings.Split(commandText, " "))
		if subCommand.ValidArgsFunction == nil {
			t.Fatalf("ValidArgsFunction was not set for command %s", commandText)
		}
	}

	// Check some flags to make sure they have a completion function
	for _, flagName := range flagsToCheck {
		err := cmd.RegisterFlagCompletionFunc(flagName, func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return nil, cobra.ShellCompDirectiveError
		})
		if err == nil || !strings.Contains(err.Error(), "already registered") {
			t.Fatalf("flag completion function was not registered for flag %s", flagName)
		}
	}
}

func TestClusterCompletion(t *testing.T) {
	fakeConfigAccess := &cmdtesting.FakeConfigAccess{
		Config: &clientcmdapi.Config{
			Clusters: map[string]*clientcmdapi.Cluster{
				"foo": {},
				"bar": {},
				"baz": {},
			},
		},
	}

	completionProvider := &clusterCompletionProvider{fakeConfigAccess}

	testCases := []struct {
		name                       string
		completionFunction         completionFunction
		args                       []string
		toComplete                 string
		expectedCompletions        []string
	}{
		{
			name:                       "should return all completions when toComplete is empty",
			completionFunction:         completionProvider.getFlagCompletions,
			args:                       []string{},
			toComplete:                 "",
			expectedCompletions:        []string{"foo", "bar", "baz"},
		},
		{
			name:                       "should return only completions that start with toComplete's value",
			completionFunction:         completionProvider.getFlagCompletions,
			args:                       []string{},
			toComplete:                 "ba",
			expectedCompletions:        []string{"bar", "baz"},
		},
		{
			name:                       "should return no completions if nothing starts with toComplete's value",
			completionFunction:         completionProvider.getFlagCompletions,
			args:                       []string{},
			toComplete:                 "zap",
			expectedCompletions:        []string{},
		},
		{
			name:                       "should return arg completions when no args have been specified yet",
			completionFunction:         completionProvider.getArgCompletions,
			args:                       []string{},
			toComplete:                 "",
			expectedCompletions:        []string{"foo", "bar", "baz"},
		},
		{
			name:                       "should return no arg completions when args have already been specified",
			completionFunction:         completionProvider.getArgCompletions,
			args:                       []string{"something"},
			toComplete:                 "",
			expectedCompletions:        nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			completions, directive := tc.completionFunction(newFakeCommand(), tc.args, tc.toComplete)
			assertCompletionsMatch(t, tc.expectedCompletions, completions)
			if expectedDirective := cobra.ShellCompDirectiveNoFileComp; expectedDirective != directive {
				t.Errorf("Expected shell comp directive: %v, but got %v", expectedDirective, directive)
			}
		})
	}
}

func TestContextCompletion(t *testing.T) {
	fakeConfigAccess := &cmdtesting.FakeConfigAccess{
		Config: &clientcmdapi.Config{
			Contexts: map[string]*clientcmdapi.Context{
				"foo": {},
				"bar": {},
				"baz": {},
			},
		},
	}

	completionProvider := &contextCompletionProvider{fakeConfigAccess}

	testCases := []struct {
		name                       string
		completionFunction         completionFunction
		args                       []string
		toComplete                 string
		expectedCompletions        []string
	}{
		{
			name:                       "should return all flag completions when toComplete is empty",
			completionFunction:         completionProvider.getFlagCompletions,
			args:                       []string{},
			toComplete:                 "",
			expectedCompletions:        []string{"foo", "bar", "baz"},
		},
		{
			name:                       "should return only flag completions that start with toComplete's value",
			completionFunction:         completionProvider.getFlagCompletions,
			args:                       []string{},
			toComplete:                 "ba",
			expectedCompletions:        []string{"bar", "baz"},
		},
		{
			name:                       "should return no flag completions if nothing starts with toComplete's value",
			completionFunction:         completionProvider.getFlagCompletions,
			args:                       []string{},
			toComplete:                 "zap",
			expectedCompletions:        []string{},
		},
		{
			name:                       "should return arg completions when no args have been specified yet",
			completionFunction:         completionProvider.getArgCompletions,
			args:                       []string{},
			toComplete:                 "",
			expectedCompletions:        []string{"foo", "bar", "baz"},
		},
		{
			name:                       "should return no arg completions when args have already been specified",
			completionFunction:         completionProvider.getArgCompletions,
			args:                       []string{"something"},
			toComplete:                 "",
			expectedCompletions:        nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			completions, directive := tc.completionFunction(newFakeCommand(), tc.args, tc.toComplete)
			assertCompletionsMatch(t, tc.expectedCompletions, completions)
			if expectedDirective := cobra.ShellCompDirectiveNoFileComp; expectedDirective != directive {
				t.Errorf("Expected shell comp directive: %v, but got %v", expectedDirective, directive)
			}
		})
	}
}

func TestUserCompletion(t *testing.T) {
	fakeConfigAccess := &cmdtesting.FakeConfigAccess{
		Config: &clientcmdapi.Config{
			AuthInfos: map[string]*clientcmdapi.AuthInfo{
				"foo": {},
				"bar": {},
				"baz": {},
			},
		},
	}

	completionProvider := &userCompletionProvider{fakeConfigAccess}

	testCases := []struct {
		name                       string
		completionFunction         completionFunction
		toComplete                 string
		expectedCompletions        []string
	}{
		{
			name:                       "should return all flag completions when toComplete is empty",
			completionFunction:         completionProvider.getFlagCompletions,
			toComplete:                 "",
			expectedCompletions:        []string{"foo", "bar", "baz"},
		},
		{
			name:                       "should return only flag completions that start with toComplete's value",
			completionFunction:         completionProvider.getFlagCompletions,
			toComplete:                 "ba",
			expectedCompletions:        []string{"bar", "baz"},
		},
		{
			name:                       "should return no flag completions if nothing starts with toComplete's value",
			completionFunction:         completionProvider.getFlagCompletions,
			toComplete:                 "zap",
			expectedCompletions:        []string{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			completions, directive := tc.completionFunction(newFakeCommand(), []string{}, tc.toComplete)
			assertCompletionsMatch(t, tc.expectedCompletions, completions)
			if expectedDirective := cobra.ShellCompDirectiveNoFileComp; expectedDirective != directive {
				t.Errorf("Expected shell comp directive: %v, but got %v", expectedDirective, directive)
			}
		})
	}
}

func TestNamespaceCompletion(t *testing.T) {
	tf := cmdtesting.NewTestFactory()
	defer tf.Cleanup()

	codec := scheme.Codecs.LegacyCodec(scheme.Scheme.PrioritizedVersionsAllGroups()...)
	tf.UnstructuredClient = &fake.RESTClient{
		NegotiatedSerializer: resource.UnstructuredPlusDefaultContentConfig().NegotiatedSerializer,
		Client: fake.CreateHTTPClient(func(req *http.Request) (*http.Response, error) {
			switch method, path := req.Method, req.URL.Path; {
			case method == "GET" && path == "/namespaces":
				return &http.Response{
					StatusCode: http.StatusOK,
					Header:     cmdtesting.DefaultHeader(),
					Body: cmdtesting.ObjBody(codec, &v1.NamespaceList{
						Items: []v1.Namespace{
							{ObjectMeta: metav1.ObjectMeta{Name: "foo"}},
							{ObjectMeta: metav1.ObjectMeta{Name: "bar"}},
							{ObjectMeta: metav1.ObjectMeta{Name: "baz"}},
						},
					}),
				}, nil
			default:
				t.Errorf("unexpected request (Method=%s, Path=%s)", method, path)
				return nil, fmt.Errorf("unexpected request")
			}
		}),
	}

	completionProvider := &namespaceCompletionProvider{tf}

	testCases := []struct {
		name                       string
		completionFunction         completionFunction
		toComplete                 string
		expectedCompletions        []string
	}{
		{
			name:                       "should return all flag completions when toComplete is empty",
			completionFunction:         completionProvider.getFlagCompletions,
			toComplete:                 "",
			expectedCompletions:        []string{"foo", "bar", "baz"},
		},
		{
			name:                       "should return only flag completions that start with toComplete's value",
			completionFunction:         completionProvider.getFlagCompletions,
			toComplete:                 "ba",
			expectedCompletions:        []string{"bar", "baz"},
		},
		{
			name:                       "should return no flag completions if nothing starts with toComplete's value",
			completionFunction:         completionProvider.getFlagCompletions,
			toComplete:                 "zap",
			expectedCompletions:        []string{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			completions, directive := tc.completionFunction(newFakeCommand(), []string{}, tc.toComplete)
			assertCompletionsMatch(t, tc.expectedCompletions, completions)
			if expectedDirective := cobra.ShellCompDirectiveNoFileComp; expectedDirective != directive {
				t.Errorf("Expected shell comp directive: %v, but got %v", expectedDirective, directive)
			}
		})
	}
}

func TestNodeCompletion(t *testing.T) {
	tf := cmdtesting.NewTestFactory()
	defer tf.Cleanup()

	codec := scheme.Codecs.LegacyCodec(scheme.Scheme.PrioritizedVersionsAllGroups()...)
	tf.UnstructuredClient = &fake.RESTClient{
		NegotiatedSerializer: resource.UnstructuredPlusDefaultContentConfig().NegotiatedSerializer,
		Client: fake.CreateHTTPClient(func(req *http.Request) (*http.Response, error) {
			switch method, path := req.Method, req.URL.Path; {
			case method == "GET" && path == "/nodes":
				return &http.Response{
					StatusCode: http.StatusOK,
					Header:     cmdtesting.DefaultHeader(),
					Body: cmdtesting.ObjBody(codec, &v1.NodeList{
						Items: []v1.Node{
							{ObjectMeta: metav1.ObjectMeta{Name: "foo"}},
							{ObjectMeta: metav1.ObjectMeta{Name: "bar"}},
							{ObjectMeta: metav1.ObjectMeta{Name: "baz"}},
						},
					}),
				}, nil
			default:
				t.Errorf("unexpected request (Method=%s, Path=%s)", method, path)
				return nil, fmt.Errorf("unexpected request")
			}
		}),
	}

	completionProvider := &nodeCompletionProvider{tf}

	testCases := []struct {
		name                       string
		completionFunction         completionFunction
		toComplete                 string
		expectedCompletions        []string
	}{
		{
			name:                       "should return all arg completions when toComplete is empty",
			completionFunction:         completionProvider.getArgCompletions,
			toComplete:                 "",
			expectedCompletions:        []string{"foo", "bar", "baz"},
		},
		{
			name:                       "should return only arg completions that start with toComplete's value",
			completionFunction:         completionProvider.getArgCompletions,
			toComplete:                 "ba",
			expectedCompletions:        []string{"bar", "baz"},
		},
		{
			name:                       "should return no arg completions if nothing starts with toComplete's value",
			completionFunction:         completionProvider.getArgCompletions,
			toComplete:                 "zap",
			expectedCompletions:        []string{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			completions, directive := tc.completionFunction(newFakeCommand(), []string{}, tc.toComplete)
			assertCompletionsMatch(t, tc.expectedCompletions, completions)
			if expectedDirective := cobra.ShellCompDirectiveNoFileComp; expectedDirective != directive {
				t.Errorf("Expected shell comp directive: %v, but got %v", expectedDirective, directive)
			}
		})
	}
}

func TestPodCompletion(t *testing.T) {
	tf := cmdtesting.NewTestFactory()
	defer tf.Cleanup()

	codec := scheme.Codecs.LegacyCodec(scheme.Scheme.PrioritizedVersionsAllGroups()...)
	tf.UnstructuredClient = &fake.RESTClient{
		NegotiatedSerializer: resource.UnstructuredPlusDefaultContentConfig().NegotiatedSerializer,
		Client: fake.CreateHTTPClient(func(req *http.Request) (*http.Response, error) {
			switch method, path := req.Method, req.URL.Path; {
			case method == "GET" && path == "/namespaces/default/pods":
				return &http.Response{
					StatusCode: http.StatusOK,
					Header:     cmdtesting.DefaultHeader(),
					Body: cmdtesting.ObjBody(codec, &v1.PodList{
						Items: []v1.Pod{
							{ObjectMeta: metav1.ObjectMeta{Name: "foo"}},
							{ObjectMeta: metav1.ObjectMeta{Name: "bar"}},
							{ObjectMeta: metav1.ObjectMeta{Name: "baz"}},
						},
					}),
				}, nil
			case method == "GET" && path == "/namespaces/test/pods":
				return &http.Response{
					StatusCode: http.StatusOK,
					Header:     cmdtesting.DefaultHeader(),
					Body: cmdtesting.ObjBody(codec, &v1.PodList{
						Items: []v1.Pod{
							{ObjectMeta: metav1.ObjectMeta{Name: "first"}},
							{ObjectMeta: metav1.ObjectMeta{Name: "second"}},
							{ObjectMeta: metav1.ObjectMeta{Name: "third"}},
						},
					}),
				}, nil
			default:
				t.Errorf("unexpected request (Method=%s, Path=%s)", method, path)
				return nil, fmt.Errorf("unexpected request")
			}
		}),
	}

	completionProvider := &podCompletionProvider{tf}

	testCases := []struct {
		name                       string
		cmd                        *cobra.Command
		completionFunction         completionFunction
		args                       []string
		toComplete                 string
		expectedCompletions        []string
	}{
		{
			name:                       "should return all arg completions when toComplete is empty",
			cmd:                        newFakeCommand(),
			completionFunction:         completionProvider.getArgCompletions,
			args:                       []string{},
			toComplete:                 "",
			expectedCompletions:        []string{"foo", "bar", "baz"},
		},
		{
			name:                       "should return only arg completions that start with toComplete's value",
			cmd:                        newFakeCommand(),
			completionFunction:         completionProvider.getArgCompletions,
			args:                       []string{},
			toComplete:                 "ba",
			expectedCompletions:        []string{"bar", "baz"},
		},
		{
			name:                       "should return no arg completions if nothing starts with toComplete's value",
			cmd:                        newFakeCommand(),
			completionFunction:         completionProvider.getArgCompletions,
			args:                       []string{},
			toComplete:                 "zap",
			expectedCompletions:        []string{},
		},
		{
			name:                       "should return no arg completions when args have already been specified",
			cmd:                        newFakeCommand(),
			completionFunction:         completionProvider.getArgCompletions,
			args:                       []string{"something"},
			toComplete:                 "",
			expectedCompletions:        nil,
		},
		{
			name:                       "should return all arg completions for non-default namespace",
			cmd:                        newFakeCommandWithNamespace("test"),
			completionFunction:         completionProvider.getArgCompletions,
			args:                       []string{},
			toComplete:                 "",
			expectedCompletions:        []string{"first", "second", "third"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			completions, directive := tc.completionFunction(tc.cmd, tc.args, tc.toComplete)
			assertCompletionsMatch(t, tc.expectedCompletions, completions)
			if expectedDirective := cobra.ShellCompDirectiveNoFileComp; expectedDirective != directive {
				t.Errorf("Expected shell comp directive: %v, but got %v", expectedDirective, directive)
			}
		})
	}
}

func TestPodAndContainerPodCompletion(t *testing.T) {
	tf := cmdtesting.NewTestFactory()
	defer tf.Cleanup()

	codec := scheme.Codecs.LegacyCodec(scheme.Scheme.PrioritizedVersionsAllGroups()...)

	tf.UnstructuredClient = &fake.RESTClient{
		NegotiatedSerializer: resource.UnstructuredPlusDefaultContentConfig().NegotiatedSerializer,
		Client: fake.CreateHTTPClient(func(req *http.Request) (*http.Response, error) {
			switch method, path := req.Method, req.URL.Path; {
			case method == "GET" && path == "/namespaces/default/pods":
				return &http.Response{
					StatusCode: http.StatusOK,
					Header:     cmdtesting.DefaultHeader(),
					Body: cmdtesting.ObjBody(codec, &v1.PodList{
						Items: []v1.Pod{
							{ObjectMeta: metav1.ObjectMeta{Name: "foo"}},
							{ObjectMeta: metav1.ObjectMeta{Name: "bar"}},
							{ObjectMeta: metav1.ObjectMeta{Name: "baz"}},
						},
					}),
				}, nil
			case method == "GET" && path == "/namespaces/test/pods":
				return &http.Response{
					StatusCode: http.StatusOK,
					Header:     cmdtesting.DefaultHeader(),
					Body: cmdtesting.ObjBody(codec, &v1.PodList{
						Items: []v1.Pod{
							{ObjectMeta: metav1.ObjectMeta{Name: "first"}},
							{ObjectMeta: metav1.ObjectMeta{Name: "second"}},
							{ObjectMeta: metav1.ObjectMeta{Name: "third"}},
						},
					}),
				}, nil
			default:
				t.Errorf("unexpected request (method=%s, path=%s)", method, path)
				return nil, fmt.Errorf("unexpected request")
			}
		}),
	}

	completionProvider := &podAndContainerCompletionProvider{tf}

	testCases := []struct {
		name                       string
		cmd                        *cobra.Command
		completionFunction         completionFunction
		args                       []string
		toComplete                 string
		expectedCompletions        []string
	}{
		{
			name:                       "should return all pod arg completions when toComplete is empty",
			cmd:                        newFakeCommand(),
			completionFunction:         completionProvider.getArgCompletions,
			args:                       []string{},
			toComplete:                 "",
			expectedCompletions:        []string{"foo", "bar", "baz"},
		},
		{
			name:                       "should return only pod arg completions that start with toComplete's value",
			cmd:                        newFakeCommand(),
			completionFunction:         completionProvider.getArgCompletions,
			args:                       []string{},
			toComplete:                 "ba",
			expectedCompletions:        []string{"bar", "baz"},
		},
		{
			name:                       "should return no pod arg completions if nothing starts with toComplete's value",
			cmd:                        newFakeCommand(),
			completionFunction:         completionProvider.getArgCompletions,
			args:                       []string{},
			toComplete:                 "zap",
			expectedCompletions:        []string{},
		},
		{
			name:                       "should return all pod arg completions for non-default namespace",
			cmd:                        newFakeCommandWithNamespace("test"),
			completionFunction:         completionProvider.getArgCompletions,
			args:                       []string{},
			toComplete:                 "",
			expectedCompletions:        []string{"first", "second", "third"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			completions, directive := tc.completionFunction(tc.cmd, tc.args, tc.toComplete)
			assertCompletionsMatch(t, tc.expectedCompletions, completions)
			if expectedDirective := cobra.ShellCompDirectiveNoFileComp; expectedDirective != directive {
				t.Errorf("Expected shell comp directive: %v, but got %v", expectedDirective, directive)
			}
		})
	}
}

func TestPodAndContainerContainerCompletion(t *testing.T) {
	tf := cmdtesting.NewTestFactory()
	defer tf.Cleanup()

	codec := scheme.Codecs.LegacyCodec(scheme.Scheme.PrioritizedVersionsAllGroups()...)

	tf.UnstructuredClient = &fake.RESTClient{
		NegotiatedSerializer: scheme.Codecs.WithoutConversion(),
		Client: fake.CreateHTTPClient(func(req *http.Request) (*http.Response, error) {
			switch method, path := req.Method, req.URL.Path; {
			case method == "GET" && path == "/namespaces/default/pods/foo":
				return &http.Response{
					StatusCode: http.StatusOK,
					Header:     cmdtesting.DefaultHeader(),
					Body: cmdtesting.ObjBody(codec, &v1.Pod{
						ObjectMeta: metav1.ObjectMeta{Name: "foo", Namespace: "default", ResourceVersion: "1"},
						Spec: v1.PodSpec{
							Containers: []v1.Container{
								{Name: "aaa"},
								{Name: "bbb"},
							},
						},
						Status: v1.PodStatus{
							Phase: v1.PodRunning,
						},
					}),
				}, nil
			case method == "GET" && path == "/namespaces/test/pods/first":
				return &http.Response{
					StatusCode: http.StatusOK,
					Header:     cmdtesting.DefaultHeader(),
					Body: cmdtesting.ObjBody(codec, &v1.Pod{
						ObjectMeta: metav1.ObjectMeta{Name: "first", Namespace: "test", ResourceVersion: "1"},
						Spec: v1.PodSpec{
							Containers: []v1.Container{
								{Name: "ccc"},
								{Name: "ddd"},
							},
						},
						Status: v1.PodStatus{
							Phase: v1.PodRunning,
						},
					}),
				}, nil
			default:
				t.Errorf("unexpected request (method=%s, path=%s)", method, path)
				return nil, fmt.Errorf("unexpected request")
			}
		}),
	}

	completionProvider := &podAndContainerCompletionProvider{tf}

	testCases := []struct {
		name                       string
		cmd                        *cobra.Command
		completionFunction         completionFunction
		args                       []string
		toComplete                 string
		expectedCompletions        []string
	}{
		{
			name:                       "should return all container arg completions when toComplete is empty",
			cmd:                        newFakeCommand(),
			completionFunction:         completionProvider.getArgCompletions,
			args:                       []string{"foo"},
			toComplete:                 "",
			expectedCompletions:        []string{"aaa", "bbb"},
		},
		{
			name:                       "should return only container arg completions that start with toComplete's value",
			cmd:                        newFakeCommand(),
			completionFunction:         completionProvider.getArgCompletions,
			args:                       []string{"foo"},
			toComplete:                 "b",
			expectedCompletions:        []string{"bbb"},
		},
		{
			name:                       "should return no container arg completions if nothing starts with toComplete's value",
			cmd:                        newFakeCommand(),
			completionFunction:         completionProvider.getArgCompletions,
			args:                       []string{"foo"},
			toComplete:                 "c",
			expectedCompletions:        []string{},
		},
		{
			name:                       "should return all container arg completions for non-default namespace",
			cmd:                        newFakeCommandWithNamespace("test"),
			completionFunction:         completionProvider.getArgCompletions,
			args:                       []string{"first"},
			toComplete:                 "",
			expectedCompletions:        []string{"ccc", "ddd"},
		},
		{
			name:                       "should return no arg completions when both args have already been specified",
			cmd:                        newFakeCommand(),
			completionFunction:         completionProvider.getArgCompletions,
			args:                       []string{"foo", "aaa"},
			toComplete:                 "",
			expectedCompletions:        nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			completions, directive := tc.completionFunction(tc.cmd, tc.args, tc.toComplete)
			assertCompletionsMatch(t, tc.expectedCompletions, completions)
			if expectedDirective := cobra.ShellCompDirectiveNoFileComp; expectedDirective != directive {
				t.Errorf("Expected shell comp directive: %v, but got %v", expectedDirective, directive)
			}
		})
	}
}

func TestResourceTypeAndResourceCompletion(t *testing.T) {
	// TODO: Add tests for Resource Type and Resource completion
}

func TestLocalOrRemoteFileCompletion(t *testing.T) {
	// TODO: Add tests for Local or Remote File completion
}

func newFakeCommand() *cobra.Command {
	cmd := &cobra.Command{}
	cmd.Flags().String("namespace", "", "")
	return cmd
}

func newFakeCommandWithNamespace(namespace string) *cobra.Command {
	cmd := &cobra.Command{}
	cmd.Flags().String("namespace", namespace, "")
	return cmd
}

func assertCompletionsMatch(t *testing.T, expectedCompletions, completions []string) {
	for _, c := range completions {
		found := false
		for _, ec := range expectedCompletions {
			if c == ec {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Found completion %q that was not in the expected completions: %v", c, expectedCompletions)
		}
	}
	for _, ec := range expectedCompletions {
		found := false
		for _, c := range completions {
			if c == ec {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Completion %q was expected but not found in the completions: %v", ec, completions)
		}
	}
}
