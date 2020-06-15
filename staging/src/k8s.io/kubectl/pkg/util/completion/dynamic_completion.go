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
	"github.com/spf13/cobra"
	"io/ioutil"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
	"k8s.io/kubectl/pkg/scheme"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

type argCompletionProvider interface {
	getArgCompletions(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective)
}

type flagCompletionProvider interface {
	getFlagCompletions(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective)
}

// RegisterDynamicCompletions registers all arg and flag completions
func RegisterDynamicCompletions(cmd *cobra.Command, f cmdutil.Factory, configAccess clientcmd.ConfigAccess) {
	clusterCompletion                 := &clusterCompletionProvider{configAccess}
	contextCompletion                 := &contextCompletionProvider{configAccess}
	localOrRemoteFileCompletion       := &localOrRemoteFileCompletionProvider{f}
	namespaceCompletion               := &namespaceCompletionProvider{f}
	nodeCompletion                    := &nodeCompletionProvider{f}
	podAndContainerCompletion         := &podAndContainerCompletionProvider{f}
	podCompletion                     := &podCompletionProvider{f}
	resourceTypeAndResourceCompletion := &resourceTypeAndResourceCompletionProvider{f}
	userCompletion                    := &userCompletionProvider{configAccess}

	// Arg Completions
	registerArgCompletion(cmd, "annotate", resourceTypeAndResourceCompletion)
	registerArgCompletion(cmd, "apply edit-last-applied", resourceTypeAndResourceCompletion)
	registerArgCompletion(cmd, "apply view-last-applied", resourceTypeAndResourceCompletion)
	registerArgCompletion(cmd, "attach", podCompletion)
	registerArgCompletion(cmd, "config rename-context", contextCompletion)
	registerArgCompletion(cmd, "config use-context", contextCompletion)
	registerArgCompletion(cmd, "cordon", nodeCompletion)
	registerArgCompletion(cmd, "cp", localOrRemoteFileCompletion)
	registerArgCompletion(cmd, "delete", resourceTypeAndResourceCompletion)
	registerArgCompletion(cmd, "delete cluster", clusterCompletion)
	registerArgCompletion(cmd, "describe", resourceTypeAndResourceCompletion)
	registerArgCompletion(cmd, "drain", nodeCompletion)
	registerArgCompletion(cmd, "edit", resourceTypeAndResourceCompletion)
	registerArgCompletion(cmd, "exec", podCompletion)
	registerArgCompletion(cmd, "get", resourceTypeAndResourceCompletion)
	registerArgCompletion(cmd, "label", resourceTypeAndResourceCompletion)
	registerArgCompletion(cmd, "logs", podAndContainerCompletion)
	registerArgCompletion(cmd, "patch", resourceTypeAndResourceCompletion)
	registerArgCompletion(cmd, "port-forward", podCompletion)
	registerArgCompletion(cmd, "rollout", resourceTypeAndResourceCompletion)
	registerArgCompletion(cmd, "taint", resourceTypeAndResourceCompletion)
	registerArgCompletion(cmd, "top node", nodeCompletion)
	registerArgCompletion(cmd, "top pod ", podCompletion)
	registerArgCompletion(cmd, "uncordon", nodeCompletion)

	// Flag Completions
	registerFlagCompletion(cmd, "cluster", clusterCompletion)
	registerFlagCompletion(cmd, "context", contextCompletion)
	registerFlagCompletion(cmd, "namespace", namespaceCompletion)
	registerFlagCompletion(cmd, "user", userCompletion)
}

func registerArgCompletion(cmd *cobra.Command, commandText string, completionProvider argCompletionProvider) {
	subCommand, _, err := cmd.Find(strings.Split(commandText, " "))
	if err != nil {
		cobra.CompError(fmt.Sprintf("Error finding %s subcommand: %v", commandText, err))
		return
	}
	subCommand.ValidArgsFunction = completionProvider.getArgCompletions
}

func registerFlagCompletion(cmd *cobra.Command, flagName string, completionProvider flagCompletionProvider) {
	err := cmd.RegisterFlagCompletionFunc(flagName, completionProvider.getFlagCompletions)
	if err != nil {
		cobra.CompError(fmt.Sprintf("Failed to register flag completion function for %s: %v", flagName, err))
	}
}

type clusterCompletionProvider struct{
	configAccess clientcmd.ConfigAccess
}

func (p *clusterCompletionProvider) getArgCompletions(_ *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) > 0 {
		// Arg has already been completed, so return an empty completion
		return completionResult(nil, nil)
	}
	return completionResult(getClusterNames(p.configAccess, toComplete))
}

func (p *clusterCompletionProvider) getFlagCompletions(_ *cobra.Command, _ []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return completionResult(getClusterNames(p.configAccess, toComplete))
}

type contextCompletionProvider struct{
	configAccess clientcmd.ConfigAccess
}

func (p *contextCompletionProvider) getArgCompletions(_ *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) > 0 {
		// Arg has already been completed, so return an empty completion
		return completionResult(nil, nil)
	}
	return completionResult(getContextNames(p.configAccess, toComplete))
}

func (p *contextCompletionProvider) getFlagCompletions(_ *cobra.Command, _ []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return completionResult(getContextNames(p.configAccess, toComplete))
}

type namespaceCompletionProvider struct{
	factory cmdutil.Factory
}

func (p *namespaceCompletionProvider) getFlagCompletions(_ *cobra.Command, _ []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return completionResult(getResourceNames(p.factory, "", "namespaces", toComplete))
}

type nodeCompletionProvider struct{
	factory cmdutil.Factory
}

func (p *nodeCompletionProvider) getArgCompletions(_ *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) > 0 {
		// Arg has already been completed, so return an empty completion
		return completionResult(nil, nil)
	}
	return completionResult(getResourceNames(p.factory, "", "nodes", toComplete))
}

type podCompletionProvider struct{
	factory cmdutil.Factory
}

func (p *podCompletionProvider) getArgCompletions(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) > 0 {
		// Arg has already been completed, so return an empty completion
		return completionResult(nil, nil)
	}

	if isAllNamespacesFlagSet(cmd) {
		// This completion isn't valid if all-namespaces flag is set, so return an empty completion
		return completionResult(nil, nil)
	}

	return completionResult(getResourceNames(p.factory, getCommandNamespace(cmd), "pods", toComplete))
}

type podAndContainerCompletionProvider struct{
	factory cmdutil.Factory
}

func (p *podAndContainerCompletionProvider) getArgCompletions(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if isAllNamespacesFlagSet(cmd) {
		// This completion isn't valid if all-namespaces flag is set, so return an empty completion
		return completionResult(nil, nil)
	}

	// If no args are specified yet, return pod name completion
	if len(args) == 0 {
		return completionResult(getResourceNames(p.factory, getCommandNamespace(cmd), "pods", toComplete))
	}

	// If one arg has already been specified, use container name completion with the first arg as the pod name
	if len(args) == 1 {
		podName := args[0]
		return completionResult(getContainerNames(p.factory, getCommandNamespace(cmd), podName, toComplete))
	}

	// Both args have already been completed, so return an empty completion
	return completionResult(nil, nil)
}

type resourceTypeAndResourceCompletionProvider struct{
	factory cmdutil.Factory
}

func (p *resourceTypeAndResourceCompletionProvider) getArgCompletions(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if isAllNamespacesFlagSet(cmd) {
		// This completion isn't valid if all-namespaces flag is set, so return an empty completion
		return completionResult(nil, nil)
	}

	// If no args are specified yet, use resource type completion
	if len(args) == 0 {
		return completionResult(getResourceTypes(p.factory, toComplete))
	}

	// If one arg has already been specified, use resource name completion with the first arg as the resource type
	if len(args) == 1 {
		resourceType := args[0]
		return completionResult(getResourceNames(p.factory, getCommandNamespace(cmd), resourceType, toComplete))
	}

	// Both args have already been completed, so return an empty completion
	return completionResult(nil, nil)
}

type userCompletionProvider struct{
	configAccess clientcmd.ConfigAccess
}

func (p *userCompletionProvider) getFlagCompletions(_ *cobra.Command, _ []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return completionResult(getUsers(p.configAccess, toComplete))
}

type localOrRemoteFileCompletionProvider struct{
	factory cmdutil.Factory
}

func (p *localOrRemoteFileCompletionProvider) getArgCompletions(cmd *cobra.Command, _ []string, toComplete string) ([]string, cobra.ShellCompDirective) {
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
		resourceNames, _ := getResourceNames(p.factory, namespace, "pods", toComplete)
		result := make([]string, len(resourceNames))
		for _, name := range resourceNames {
			result = append(result, fmt.Sprintf("%s/%s", namespace, name))
		}
		return result, cobra.ShellCompDirectiveNoSpace | cobra.ShellCompDirectiveNoFileComp
	}

	// Otherwise, return a combined list of namespaces, pods, files, and directories
	var result []string
	namespaces, _ := getResourceNames(p.factory, "", "namespaces", toComplete)
	for _, namespace := range namespaces {
		result = append(result, namespace+"/")
	}
	pods, _ := getResourceNames(p.factory, getCommandNamespace(cmd), "pods", toComplete)
	for _, podName := range pods {
		result = append(result, podName+":")
	}
	filesAndDirectories, _ := getFilesAndDirectories(".")
	result = append(result, filesAndDirectories...)
	return result, cobra.ShellCompDirectiveNoSpace
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

func getClusterNames(configAccess clientcmd.ConfigAccess, prefixToMatch string) ([]string, error) {
	config, err := configAccess.GetStartingConfig()
	if err != nil {
		return nil, err
	}

	var result []string
	for clusterName := range config.Clusters {
		if len(prefixToMatch) > 0 && !strings.HasPrefix(clusterName, prefixToMatch) {
			continue
		}
		result = append(result, clusterName)
	}

	return result, nil
}

func getContainerNames(f cmdutil.Factory, namespace, podName, prefixToMatch string) ([]string, error) {
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
			if len(prefixToMatch) > 0 && !strings.HasPrefix(container.Name, prefixToMatch) {
				continue
			}
			result = append(result, container.Name)
		}
		for _, container := range pod.Spec.Containers {
			if len(prefixToMatch) > 0 && !strings.HasPrefix(container.Name, prefixToMatch) {
				continue
			}
			result = append(result, container.Name)
		}
	}
	return result, nil
}

func getContextNames(configAccess clientcmd.ConfigAccess, prefixToMatch string) ([]string, error) {
	config, err := configAccess.GetStartingConfig()
	if err != nil {
		return nil, err
	}

	var result []string
	for name := range config.Contexts {
		if len(prefixToMatch) > 0 && !strings.HasPrefix(name, prefixToMatch) {
			continue
		}
		result = append(result, name)
	}

	return result, nil
}

func getResourceNames(f cmdutil.Factory, namespace, resourceType, prefixToMatch string) ([]string, error) {
	infos, err := f.NewBuilder().
		Unstructured().
		NamespaceParam(namespace).DefaultNamespace().
		ResourceTypeOrNameArgs(true, resourceType).
		ContinueOnError().
		Flatten().
		TransformRequests(func (req *rest.Request) { req.Timeout(5*time.Second) }).
		Do().
		Infos()
	if err != nil {
		return nil, err
	}

	var names []string
	for _, info := range infos {
		if len(prefixToMatch) > 0 && !strings.HasPrefix(info.Name, prefixToMatch) {
			continue
		}
		names = append(names, info.Name)
	}

	return names, nil
}

func getResourceTypes(f cmdutil.Factory, prefixToMatch string) ([]string, error) {
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

func getUsers(configAccess clientcmd.ConfigAccess, prefixToMatch string) ([]string, error) {
	config, err := configAccess.GetStartingConfig()
	if err != nil {
		return nil, err
	}

	var result []string
	for userName := range config.AuthInfos {
		if len(prefixToMatch) > 0 && !strings.HasPrefix(userName, prefixToMatch) {
			continue
		}
		result = append(result, userName)
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
