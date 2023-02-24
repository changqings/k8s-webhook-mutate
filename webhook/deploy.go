package webhook

import (
	"encoding/json"
	"fmt"
	"io"
	"k8s-webhook-test/common"
	"log"
	"net/http"

	v1 "k8s.io/api/admission/v1"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Deploy struct {
	Name          string
	Namespace     string
	PodNamePrefix string
}

var someAnnoMap map[string]string = map[string]string{
	"k8s-webhook-test": "added",
}

func (d Deploy) AddAnno(w http.ResponseWriter, r *http.Request) {

	// some check
	if r.Header.Get("Content-Type") != "application/json" {
		sendError(fmt.Errorf("request content-type=%s, not equal application/json", r.Header.Get("Content-Type")), w)
		return
	}

	// check not modify

	body, err := io.ReadAll(r.Body)
	if err != nil {
		sendError(err, w)
		return
	}

	//
	ar := v1.AdmissionReview{}
	if err := json.Unmarshal(body, &ar); err != nil {
		sendError(err, w)
		return
	}

	arReq := ar.Request
	arRes := ar.Response

	if arReq == nil {
		sendError(fmt.Errorf("arReq == nil"), w)
		return
	}

	// try to get deployment object, and modify it
	var deploy *appsv1.Deployment
	if err := json.Unmarshal(arReq.Object.Raw, &deploy); err != nil {
		sendError(err, w)
		return
	}

	if !(deploy.Namespace == d.Namespace && deploy.Name == d.Name) {
		log.Printf("deploy %s/%s not match %s/%s, skip webhook mutation", deploy.Namespace, deploy.Name, d.Namespace, d.Name)
		return
	}

	//
	var patchDeployAnno *common.Patch
	hasAnnos := len(deploy.ObjectMeta.Annotations) > 0
	for k, v := range someAnnoMap {
		if !hasAnnos {
			patchDeployAnno = &common.Patch{
				OP:    "add",
				Path:  "/metadata/annotations",
				Value: someAnnoMap,
			}
		} else {
			patchDeployAnno = &common.Patch{
				OP:    "add",
				Path:  fmt.Sprintf("/metadata/annotations/%s", k),
				Value: v,
			}
		}

	}
	patchDeployAnno = &common.Patch{
		OP:    "add",
		Path:  "/metadata/annotations",
		Value: someAnnoMap,
	}

	patchDeployByte, err := json.Marshal(patchDeployAnno)
	if err != nil {
		sendError(err, w)
		return
	}

	// update arRes
	arRes.Allowed = true
	arRes.UID = ar.Request.UID
	*arRes.PatchType = v1.PatchTypeJSONPatch
	arRes.Patch = patchDeployByte

	arRes.Result = &metav1.Status{
		Status: metav1.StatusSuccess,
	}

	// update arRes to w
	responBody, err := json.Marshal(arRes)
	if err != nil {
		sendError(err, w)
		return
	}

	if _, err := w.Write(responBody); err != nil {
		sendError(err, w)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func sendError(err error, w http.ResponseWriter) {
	log.Println(err)
	w.WriteHeader(http.StatusInternalServerError)
	fmt.Fprintf(w, "%s", err)
}
