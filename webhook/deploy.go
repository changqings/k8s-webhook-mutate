package webhook

import (
	"encoding/json"
	"fmt"
	"io"
	"k8s-webhook-mutate/common"
	"log"
	"net/http"

	admissionv1 "k8s.io/api/admission/v1"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
)

var (
	runtimeScheme = runtime.NewScheme()
	codecs        = serializer.NewCodecFactory(runtimeScheme)
	deserializer  = codecs.UniversalDeserializer()
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

	body, err := io.ReadAll(r.Body)
	if err != nil {
		sendError(err, w)
		return
	}

	//
	var ar admissionv1.AdmissionReview
	if _, _, err := deserializer.Decode(body, nil, &ar); err != nil {
		sendError(err, w)
		return
	}

	if ar.Request == nil {
		sendError(fmt.Errorf("ar.Request == nil"), w)
		return
	}

	// ar response handler logic in this func
	// create a new adminssionReview for response
	reviewRes := d.Deployment(&ar)
	resReview := admissionv1.AdmissionReview{
		TypeMeta: metav1.TypeMeta{
			APIVersion: ar.APIVersion,
			Kind:       ar.Kind,
		},
	}
	resReview.Response = reviewRes
	// resReview.Request = ar.Request

	// rewrite resReview back to webhook respon
	responBody, err := json.Marshal(resReview)
	if err != nil {
		sendError(err, w)
		return
	}

	// println res Body to check status
	// log.Printf("response body = %s", string(responBody))
	if _, err := w.Write(responBody); err != nil {
		sendError(err, w)
		return
	}
	log.Printf("Exec mutation webhook success")
}

func sendError(err error, w http.ResponseWriter) {
	log.Println(err)
	w.WriteHeader(http.StatusInternalServerError)
	fmt.Fprintf(w, "%s", err)
}

func (d Deploy) Deployment(ar *admissionv1.AdmissionReview) *admissionv1.AdmissionResponse {
	// try to get deployment object, and modify it
	var deploy *appsv1.Deployment
	if err := json.Unmarshal(ar.Request.Object.Raw, &deploy); err != nil {
		log.Printf("Unmarshal ar.Request.Object.Raw to deploy err:%v\n", err)
		ar.Response = &admissionv1.AdmissionResponse{
			Allowed: true,
			UID:     ar.Request.UID,
		}
		return ar.Response
	}
	// if check no need to update, do noting of ar.respon, else patch it with jsonPatch
	if !(deploy.Namespace == d.Namespace && deploy.Name == d.Name) {
		log.Printf("deploy %s/%s not match %s/%s, skip webhook mutation, skip update", deploy.Namespace, deploy.Name, d.Namespace, d.Name)
		ar.Response = &admissionv1.AdmissionResponse{
			Allowed: true,
			UID:     ar.Request.UID,
		}
		return ar.Response
	}
	log.Printf("deployment %s/%s exec mutation webhook ...", deploy.Namespace, deploy.Name)

	var patchDeployAnnos []common.Patch
	var patchDeployAnno common.Patch
	hasAnnos := len(deploy.ObjectMeta.Annotations) > 0
	for k, v := range someAnnoMap {
		if !hasAnnos {
			patchDeployAnno = common.Patch{
				OP:    "add",
				Path:  "/metadata/annotations",
				Value: someAnnoMap,
			}
		} else {
			patchDeployAnno = common.Patch{
				OP:    "add",
				Path:  fmt.Sprintf("/metadata/annotations/%s", k),
				Value: v,
			}
		}
		patchDeployAnnos = append(patchDeployAnnos, patchDeployAnno)

	}

	patchDeployByte, err := json.Marshal(patchDeployAnnos)
	if err != nil {
		ar.Response = &admissionv1.AdmissionResponse{
			Result: &metav1.Status{
				Message: err.Error(),
			},
		}
		return ar.Response
	}

	// update arRes
	ar.Response = &admissionv1.AdmissionResponse{
		Allowed: true,
		UID:     ar.Request.UID,
		Patch:   patchDeployByte,
		PatchType: func() *admissionv1.PatchType {
			pathType := admissionv1.PatchTypeJSONPatch
			return &pathType
		}(),
	}

	return ar.Response
}
