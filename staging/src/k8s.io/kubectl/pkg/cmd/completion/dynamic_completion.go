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

package completion

import (
	"fmt"
	"io/ioutil"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"k8s.io/kubectl/pkg/cmd/util"
)

func RegisterDynamicCompletions(cmd *cobra.Command, f util.Factory) {
	registerFlagCompletion(cmd, "namespace", getNamespaceCompletionFunc(f))
	registerFlagCompletion(cmd, "context", getContextCompletionFunc(f))
	registerFlagCompletion(cmd, "cluster", getClusterCompletionFunc(f))
	registerFlagCompletion(cmd, "user", getUserCompletionFunc(f))

	registerArgCompletion(cmd, "annotate", getResourceTypeAndResourceCompletionFunc(f))
	registerArgCompletion(cmd, "apply edit-last-applied", getResourceTypeAndResourceCompletionFunc(f))
	registerArgCompletion(cmd, "apply view-last-applied", getResourceTypeAndResourceCompletionFunc(f))
	registerArgCompletion(cmd, "attach", getPodCompletionFunc(f))
	registerArgCompletion(cmd, "config rename-context", getContextCompletionFunc(f))
	registerArgCompletion(cmd, "config use-context", getContextCompletionFunc(f))
	registerArgCompletion(cmd, "cordon", getNodeCompletionFunc(f))
	registerArgCompletion(cmd, "cp", getLocalOrRemoteFileCompletionFunc(f))
	registerArgCompletion(cmd, "delete", getResourceTypeAndResourceCompletionFunc(f))
	registerArgCompletion(cmd, "delete cluster", getClusterCompletionFunc(f))
	registerArgCompletion(cmd, "describe", getResourceTypeAndResourceCompletionFunc(f))
	registerArgCompletion(cmd, "drain", getNodeCompletionFunc(f))
	registerArgCompletion(cmd, "edit", getResourceTypeAndResourceCompletionFunc(f))
	registerArgCompletion(cmd, "exec", getPodCompletionFunc(f))
	registerArgCompletion(cmd, "get", getResourceTypeAndResourceCompletionFunc(f))
	registerArgCompletion(cmd, "label", getResourceTypeAndResourceCompletionFunc(f))
	registerArgCompletion(cmd, "logs", getPodAndContainerCompletionFunc(f))
	registerArgCompletion(cmd, "patch", getResourceTypeAndResourceCompletionFunc(f))
	registerArgCompletion(cmd, "port-forward", getPodCompletionFunc(f))
	registerArgCompletion(cmd, "rollout", getResourceTypeAndResourceCompletionFunc(f))
	registerArgCompletion(cmd, "taint", getResourceTypeAndResourceCompletionFunc(f))
	registerArgCompletion(cmd, "top node", getNodeCompletionFunc(f))
	registerArgCompletion(cmd, "top pod ", getPodCompletionFunc(f))
	registerArgCompletion(cmd, "uncordon", getNodeCompletionFunc(f))
}

type completionFunc func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective)

func registerArgCompletion(cmd *cobra.Command, commandText string, f completionFunc) {
	subCommand, _, err := cmd.Find(strings.Split(commandText, " "))
	if err != nil {
		cobra.CompError(fmt.Sprintf("Error finding %s subcommand: %v", commandText, err))
		return
	}
	subCommand.ValidArgsFunction = f
}

func registerFlagCompletion(cmd *cobra.Command, flagName string, f completionFunc) {
	err := cmd.RegisterFlagCompletionFunc(flagName, f)
	if err != nil {
		cobra.CompError(fmt.Sprintf("Failed to register flag completion function for %s: %v", flagName, err))
		return
	}
}

func getContextCompletionFunc(restClientGetter genericclioptions.RESTClientGetter,) completionFunc {
	return func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		config, err := restClientGetter.ToRawKubeConfigLoader().RawConfig()
		if err != nil {
			return completionResult(nil, err)
		}
		var result []string
		for name := range config.Contexts {
			if strings.HasPrefix(name, toComplete) {
				result = append(result, name)
			}
		}
		return completionResult(result, nil)
	}
}

func getClusterCompletionFunc(restClientGetter genericclioptions.RESTClientGetter,) completionFunc {
	return func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		config, err := restClientGetter.ToRawKubeConfigLoader().RawConfig()
		if err != nil {
			return completionResult(nil, err)
		}
		var result []string
		for name := range config.Clusters {
			if strings.HasPrefix(name, toComplete) {
				result = append(result, name)
			}
		}
		return completionResult(result, nil)
	}
}

func getUserCompletionFunc(restClientGetter genericclioptions.RESTClientGetter) completionFunc {
	return func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		config, err := restClientGetter.ToRawKubeConfigLoader().RawConfig()
		if err != nil {
			return completionResult(nil, err)
		}
		var result []string
		for name := range config.AuthInfos {
			if strings.HasPrefix(name, toComplete) {
				result = append(result, name)
			}
		}
		return completionResult(result, nil)
	}
}

func getNamespaceCompletionFunc(f util.Factory) completionFunc {
	return func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return completionResult(getResourceNames(f, "", "namespaces", toComplete))
	}
}

func getNodeCompletionFunc(f util.Factory) completionFunc {
	return func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return completionResult(getResourceNames(f, "", "nodes", toComplete))
	}
}

func getPodCompletionFunc(f util.Factory) completionFunc {
	return func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if isAllNamespacesFlagSet(cmd) {
			// This completion isn't valid if all-namespaces flag is set, so return an empty completion
			return completionResult(nil, nil)
		}
		return completionResult(getResourceNames(f, getCommandNamespace(cmd), "pods", toComplete))
	}
}

func getResourceTypeAndResourceCompletionFunc(f util.Factory) completionFunc {
	return func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if isAllNamespacesFlagSet(cmd) {
			// This completion isn't valid if all-namespaces flag is set, so return an empty completion
			return completionResult(nil, nil)
		}
		// If no args are specified yet, use resource type completion
		if len(args) == 0 {
			return completionResult(getResourceTypes(f, toComplete))
		}
		// If one arg has already been specified, use resource name completion with the first arg as the resource type
		if len(args) == 1 {
			resourceType := args[0]
			return completionResult(getResourceNames(f, getCommandNamespace(cmd), resourceType, toComplete))
		}
		// Both args have already been completed, so return an empty completion
		return completionResult(nil, nil)
	}
}

func getPodAndContainerCompletionFunc(f util.Factory) completionFunc {
	return func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if isAllNamespacesFlagSet(cmd) {
			// This completion isn't valid if all-namespaces flag is set, so return an empty completion
			return completionResult(nil, nil)
		}
		// If no args are specified yet, return pod name completion
		if len(args) == 0 {
			return completionResult(getResourceNames(f, getCommandNamespace(cmd), "pods", toComplete))
		}
		// If one arg has already been specified, use container name completion with the first arg as the pod name
		if len(args) == 1 {
			podName := args[0]
			return completionResult(getContainerNames(f, getCommandNamespace(cmd), podName, toComplete))
		}
		// Both args have already been completed, so return an empty completion
		return completionResult(nil, nil)
	}
}

func getLocalOrRemoteFileCompletionFunc(f util.Factory) completionFunc {
	return func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		// If toComplete looks like a remote path, return an empty completion
		if isRemotePathPatternMatch(toComplete) {
			return completionResult(nil, nil)
		}
		// If toComplete looks like a local path or matches an existing file or directory, return the default completion
		if isLocalPathPatternMatch(toComplete) || isExistingLocalFileOrDirectory(toComplete) {
			return nil, cobra.ShellCompDirectiveDefault
		}
		// If toComplete looks like <namespace>/<pod>, return a list of namespace/pods
		if isNamespacePodPatternMatch(toComplete) {
			parts := strings.SplitN(toComplete, "/", 2)
			namespace := parts[0]
			toComplete = parts[1]
			resourceNames, _ := getResourceNames(f, namespace, "pods", toComplete)
			result := make([]string, len(resourceNames))
			for _, name := range resourceNames {
				result = append(result, fmt.Sprintf("%s/%s", namespace, name))
			}
			return result, cobra.ShellCompDirectiveNoSpace | cobra.ShellCompDirectiveNoFileComp
		}
		// Otherwise, return a combined list of namespaces, pods, files, and directories
		var result []string
		namespaces, _ := getResourceNames(f, "", "namespaces", toComplete)
		for _, namespace := range namespaces {
			result = append(result, namespace+"/")
		}
		pods, _ := getResourceNames(f, getCommandNamespace(cmd), "pods", toComplete)
		for _, podName := range pods {
			result = append(result, podName+":")
		}
		filesAndDirectories, _ := getFilesAndDirectories(".")
		result = append(result, filesAndDirectories...)
		return result, cobra.ShellCompDirectiveNoSpace
	}
}

func isRemotePathPatternMatch(s string) bool {
	isMatch, err := regexp.MatchString("^.*:", s)
	if err != nil {
		cobra.CompError(fmt.Sprintf("isRemotePathPatternMatch Regexp error: %v", err))
		return false
	}
	return isMatch
}

func isLocalPathPatternMatch(s string) bool {
	isMatch, err := regexp.MatchString("^[/.~]", s)
	if err != nil {
		cobra.CompError(fmt.Sprintf("isLocalPathPatternMatch Regexp error: %v", err))
		return false
	}
	return isMatch
}

func isExistingLocalFileOrDirectory(s string) bool {
	fileInfo, _ := os.Stat(s)
	if fileInfo != nil {
		return true
	}
	dir := filepath.Dir(s)
	if dir != "." {
		fileInfo, _ := os.Stat(dir)
		if fileInfo != nil {
			return true
		}
	}
	return false
}

func isNamespacePodPatternMatch(s string) bool {
	isMatch, err := regexp.MatchString("^.+/", s)
	if err != nil {
		cobra.CompError(fmt.Sprintf("isNamespacePodPatternMatch Regexp error: %v", err))
		return false
	}
	return isMatch
}

func getContainerNames(f util.Factory, namespace, podName, prefixToMatch string) ([]string, error) {
	obj, err := f.NewBuilder().
		WithScheme(scheme.Scheme, scheme.Scheme.PrioritizedVersionsAllGroups()...).
		NamespaceParam(namespace).DefaultNamespace().
		ResourceNames("pods", podName).
		Do().
		Object()
	if err != nil {
		return nil, err
	}

	var result []string
	if pod, ok := obj.(*corev1.Pod); ok {
		for _, container := range pod.Spec.InitContainers {
			if strings.HasPrefix(container.Name, prefixToMatch) {
				result = append(result, container.Name)
			}
		}
		for _, container := range pod.Spec.Containers {
			if strings.HasPrefix(container.Name, prefixToMatch) {
				result = append(result, container.Name)
			}
		}
	}
	return result, nil
}

func getResourceNames(f util.Factory, namespace, resourceType, prefixToMatch string) ([]string, error) {
	infos, err := f.NewBuilder().
		Unstructured().
		NamespaceParam(namespace).DefaultNamespace().
		ResourceTypeOrNameArgs(true, resourceType).
		ContinueOnError().
		Flatten().
		TransformRequests(func(req *rest.Request) { req.Timeout(5 * time.Second) }).
		Do().
		Infos()
	if err != nil {
		return nil, err
	}

	var names []string
	for _, info := range infos {
		if strings.HasPrefix(info.Name, prefixToMatch) {
			names = append(names, info.Name)
		}
	}

	return names, nil
}

func getResourceTypes(f util.Factory, prefixToMatch string) ([]string, error) {
	discoveryClient, err := f.ToDiscoveryClient()
	if err != nil {
		return nil, err
	}
	discoveryClient.Invalidate()

	lists, err := discoveryClient.ServerPreferredResources()
	if err != nil {
		return nil, err
	}

	var result []string
	for _, list := range lists {
		gv, err := schema.ParseGroupVersion(list.GroupVersion)
		if err != nil {
			continue
		}
		for _, resource := range list.APIResources {
			if len(resource.Verbs) == 0 {
				continue
			}

			var qualifiedName string
			if len(gv.Group) > 0 {
				qualifiedName = fmt.Sprintf("%s.%s", resource.Name, gv.Group)
			} else {
				qualifiedName = resource.Name
			}

			if len(prefixToMatch) > 0 {
				var shortNameMatchesPrefix = false
				for _, shortName := range resource.ShortNames {
					if strings.HasPrefix(shortName, prefixToMatch) {
						shortNameMatchesPrefix = true
					}
				}
				if !shortNameMatchesPrefix && !strings.HasPrefix(qualifiedName, prefixToMatch) {
					continue
				}
			}
			result = append(result, qualifiedName)
		}
	}
	return result, nil
}

func getFilesAndDirectories(path string) ([]string, error) {
	fileInfos, err := ioutil.ReadDir(path)
	if err != nil {
		return nil, err
	}
	result := make([]string, 0, len(fileInfos))
	for _, fileInfo := range fileInfos {
		if fileInfo.IsDir() {
			result = append(result, fileInfo.Name()+"/")
		} else {
			result = append(result, fileInfo.Name())
		}
	}
	return result, nil
}

func isAllNamespacesFlagSet(cmd *cobra.Command) bool {
	value, _ := cmd.Flags().GetBool("all-namespaces")
	return value
}

func getCommandNamespace(cmd *cobra.Command) string {
	namespace, _ := cmd.Flags().GetString("namespace")
	if len(namespace) == 0 {
		namespace = "default"
	}
	return namespace
}

func completionResult(values []string, err error) ([]string, cobra.ShellCompDirective) {
	if err != nil {
		cobra.CompError(err.Error())
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	return values, cobra.ShellCompDirectiveNoFileComp
}
