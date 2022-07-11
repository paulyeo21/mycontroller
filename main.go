package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/wI2L/jsondiff"
	admissionv1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
)

func podAnnotatorHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// parse request
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, fmt.Sprintf("could not parse request: %v", err), http.StatusBadRequest)
			return
		}

		var aReq admissionv1.AdmissionReview

		if err := json.Unmarshal(body, &aReq); err != nil {
			http.Error(w, fmt.Sprintf("could not unmarshal req: %v", err), http.StatusBadRequest)
			return
		}

		if aReq.Request == nil {
			http.Error(w, "request is empty", http.StatusBadRequest)
			return
		}

		// mutate pod
		pod := corev1.Pod{}
		if err := json.Unmarshal(aReq.Request.Object.Raw, &pod); err != nil {
			http.Error(w, fmt.Sprintf("failed to unmarshal pod: %v", err), http.StatusBadRequest)
			return
		}

		mpod := pod.DeepCopy()
		mpod.Annotations["example-mutating-admission-webhook"] = "foo"

		patch, err := jsondiff.Compare(pod, mpod)
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to create patch: %v", err), http.StatusBadRequest)
			return
		}

		// create admission response
		patchb, err := json.Marshal(patch)
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to marshal patch: %v", err), http.StatusBadRequest)
			return
		}

		aRes := admissionv1.AdmissionReview{
			Response: &admissionv1.AdmissionResponse{
				Allowed: true,
				Patch:   patchb,
				UID:     aReq.Request.UID,
			},
		}

		bytes, err := json.Marshal(&aRes)
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to marshal response: %v", err), http.StatusInternalServerError)
			return
		}

		// return res
		w.Write(bytes)
	})
}

func main() {
	mux := http.NewServeMux()
	mux.Handle("/mutate", podAnnotatorHandler())
	server := &http.Server{
		Addr:    ":8443",
		Handler: mux,
	}
	log.Fatal(server.ListenAndServe())
}
