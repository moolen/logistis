package recorder

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/moolen/logistis/pkg/store"
	"github.com/sirupsen/logrus"
	admissionv1 "k8s.io/api/admission/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
)

func (a *Recorder) RecordEvent(request *admissionv1.AdmissionRequest) (*admissionv1.AdmissionReview, error) {
	a.store.Observe(&store.Event{
		ID:        string(request.UID),
		Group:     request.RequestKind.Group,
		Kind:      request.RequestKind.Kind,
		Namespace: request.Namespace,
		Name:      request.Name,
		Operation: string(request.Operation),
		UserInfo:  request.UserInfo.String(),
		Object:    request.Object.Raw,
		OldObject: request.OldObject.Raw,
		Timestamp: time.Now(),
	})
	return reviewResponse(request.UID, true, http.StatusAccepted, "", nil), nil
}

// CaptureEvents captures incoming events
func (a *Recorder) CaptureEvents(w http.ResponseWriter, r *http.Request) {
	logger := logrus.WithField("uri", r.RequestURI)
	logger.Debug("received validation request")

	in, err := parseRequest(*r)
	if err != nil {
		logger.Error(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	out, err := a.RecordEvent(in.Request)
	if err != nil {
		e := fmt.Sprintf("could not generate admission response: %v", err)
		logger.Error(e)
		http.Error(w, e, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	jout, err := json.Marshal(out)
	if err != nil {
		e := fmt.Sprintf("could not parse admission response: %v", err)
		logger.Error(e)
		http.Error(w, e, http.StatusInternalServerError)
		return
	}

	logger.Debug("sending response")
	logger.Debugf("%s", jout)
	fmt.Fprintf(w, "%s", jout)
}

// parseRequest extracts an AdmissionReview from an http.Request if possible
func parseRequest(r http.Request) (*admissionv1.AdmissionReview, error) {
	if r.Header.Get("Content-Type") != "application/json" {
		return nil, fmt.Errorf("Content-Type: %q should be %q",
			r.Header.Get("Content-Type"), "application/json")
	}
	bodybuf := new(bytes.Buffer)
	bodybuf.ReadFrom(r.Body)
	body := bodybuf.Bytes()
	if len(body) == 0 {
		return nil, fmt.Errorf("admission request body is empty")
	}

	var a admissionv1.AdmissionReview
	if err := json.Unmarshal(body, &a); err != nil {
		return nil, fmt.Errorf("could not parse admission review request: %v", err)
	}
	if a.Request == nil {
		return nil, fmt.Errorf("admission review can't be used: Request field is nil")
	}
	return &a, nil
}

// reviewResponse TODO: godoc
func reviewResponse(uid types.UID, allowed bool, httpCode int32,
	reason string, warnings []string) *admissionv1.AdmissionReview {
	return &admissionv1.AdmissionReview{
		TypeMeta: metav1.TypeMeta{
			Kind:       "AdmissionReview",
			APIVersion: "admission.k8s.io/v1",
		},
		Response: &admissionv1.AdmissionResponse{
			UID:      uid,
			Allowed:  allowed,
			Warnings: warnings,
			Result: &metav1.Status{
				Code:    httpCode,
				Message: reason,
			},
		},
	}
}
