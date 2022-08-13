package recorder

import (
	"encoding/json"
	"net/http"
	"strconv"
)

func (a *Recorder) ListEvents(w http.ResponseWriter, r *http.Request) {
	a.Logger.Info("received list events")
	namespace := r.URL.Query().Get("namespace")
	kind := r.URL.Query().Get("kind")
	name := r.URL.Query().Get("name")
	maxHistoryStr := r.URL.Query().Get("max-history")
	maxHistory, err := strconv.Atoi(maxHistoryStr)
	if err != nil {
		a.Logger.Errorf("unable convert max history: %s", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	a.Logger.Infof("listing for ns=%q kind=%q name=%q history=%d", namespace, kind, name, maxHistory)
	events, err := a.store.List(namespace, kind, name, maxHistory)
	if err != nil {
		a.Logger.Errorf("unable to list events: %s", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	jout, err := json.Marshal(events)
	if err != nil {
		a.Logger.Errorf("unable to marshal events: %s", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Add("content-type", "application/json")
	w.Write(jout)
}
