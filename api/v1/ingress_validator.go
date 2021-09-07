package v1

import (
	"context"
	"fmt"
	"net/http"

	"github.o-in.dwango.co.jp/naari3/ingress-sg-validator/pkg"
	networkingv1 "k8s.io/api/networking/v1"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

//+kubebuilder:webhook:path=/validate-v1-ingress,mutating=false,failurePolicy=fail,sideEffects=None,groups="networking.k8s.io";"extensions",resources=ingresses,verbs=create;update,versions=v1,name=ving.nnn.ed.nico,admissionReviewVersions={v1,v1beta1}
type IngressValidator struct {
	Client    client.Client
	decoder   *admission.Decoder
	validator pkg.SGValidator
}

func NewIngressValidator(c client.Client, v pkg.SGValidator) admission.Handler {
	return &IngressValidator{Client: c, validator: v}
}

// ingressValidator admits a ingress if a specific annotation exists.
func (a *IngressValidator) Handle(ctx context.Context, req admission.Request) admission.Response {
	ingress := &networkingv1.Ingress{}

	err := a.decoder.Decode(req, ingress)
	if err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	if err := a.validator.ValidateAnnotation(ctx, ingress); err != nil {
		return admission.Denied(fmt.Sprintf("SG Validation failed: %s", err))
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
