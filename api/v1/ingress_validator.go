package v1

import (
	"context"
	"fmt"
	"net/http"

	networkingv1 "k8s.io/api/networking/v1"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// +kubebuilder:webhook:path=/validate-v1-ingress,mutating=false,failurePolicy=fail,groups="networking.k8s.io",resources=ingresses,verbs=create;update,versions=v1,name=ving.kb.io,sideEffects=None,admissionReviewVersions={v1}
type IngressValidator struct {
	Client  client.Client
	decoder *admission.Decoder
}

func NewIngressValidator(c client.Client) admission.Handler {
	return &IngressValidator{Client: c}
}

// ingressValidator admits a ingress if a specific annotation exists.
func (a *IngressValidator) Handle(ctx context.Context, req admission.Request) admission.Response {
	ingress := &networkingv1.Ingress{}

	err := a.decoder.Decode(req, ingress)
	if err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	key := "example-mutating-admission-webhook"
	anno, found := ingress.Annotations[key]
	if !found {
		return admission.Denied(fmt.Sprintf("missing annotation %s", key))
	}
	if anno != "foo" {
		return admission.Denied(fmt.Sprintf("annotation %s did not have value %q", key, "foo"))
	}

	return admission.Allowed("")
}

// ingressValidator implements admission.DecoderInjector.
// A decoder will be automatically injected.

// InjectDecoder injects the decoder.
func (a *IngressValidator) InjectDecoder(d *admission.Decoder) error {
	a.decoder = d
	return nil
}
