/*
Copyright paskal.maksim@gmail.com
Licensed under the Apache License, Version 2.0 (the "License")
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package web

import (
	"fmt"
	"net/http"
	"os"

	"github.com/maksim-paskal/service-leader-election/pkg/api"
	"github.com/maksim-paskal/service-leader-election/pkg/config"
	log "github.com/sirupsen/logrus"
	"go.uber.org/atomic"
)

var IsMaster atomic.Bool

func StartServer() {
	log.Info(fmt.Sprintf("Starting on port %d...", *config.Port))

	err := http.ListenAndServe(fmt.Sprintf(":%d", *config.Port), GetHandler())
	if err != nil {
		log.WithError(err).Fatal()
	}
}

func GetHandler() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("/ready", handlerReady)
	mux.HandleFunc("/healthz", handlerHealthz)
	mux.HandleFunc("/debug", handlerDebug)

	return mux
}

func handlerReady(w http.ResponseWriter, r *http.Request) {
	if IsMaster.Load() {
		_, _ = w.Write([]byte("is master"))
	} else {
		w.WriteHeader(http.StatusServiceUnavailable)

		_, _ = w.Write([]byte("wait for master"))
	}
}

func handlerHealthz(w http.ResponseWriter, r *http.Request) {
	ready, container, err := api.CheckContainerIsReady(r.Context())
	if err != nil {
		log.WithError(err).Errorf("failed to check container %s ready status", container)
		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}

	// liveness probe will fail if primary container is not ready.
	if !ready {
		http.Error(w, "primary container is not ready", http.StatusInternalServerError)

		return
	}

	_, _ = w.Write([]byte("live"))
}

func handlerDebug(w http.ResponseWriter, r *http.Request) {
	_, _ = w.Write([]byte(os.Getenv("HOSTNAME")))
}
