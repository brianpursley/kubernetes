package filters

import (
	"k8s.io/apimachinery/pkg/version"
	"net/http"
)

func WithKubernetesVersion(handler http.Handler, info *version.Info) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		// Set the Kubernetes-Version header if it is not already set
		if _, ok := w.Header()["Kubernetes-Version"]; !ok {
			w.Header().Set("Kubernetes-Version", info.String())
		}
		handler.ServeHTTP(w, req)
	})
}
