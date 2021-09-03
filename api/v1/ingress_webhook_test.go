package v1

import (
	"context"
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	netv1 "k8s.io/api/networking/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func validateIngressTest(ingress *netv1.Ingress, valid bool) {
	ctx := context.Background()

	err := k8sClient.Create(ctx, ingress)

	if valid {
		Expect(err).NotTo(HaveOccurred(), "Ingress: %v", ingress)
	} else {
		Expect(err).To(HaveOccurred(), "Ingress: %v", ingress)
		statusErr := &k8serrors.StatusError{}
		Expect(errors.As(err, &statusErr)).To(BeTrue())
		expected := ingress.Annotations["message"]
		Expect(statusErr.ErrStatus.Message).To(ContainSubstring(expected))
	}
}

var _ = Describe("Ingress Webhook", func() {
	Context("validating", func() {
		const (
			IngressName      = "test-ingress"
			IngressNamespace = "default"
		)
		var prefix netv1.PathType
		prefix = "Prefix"
		validIngress := netv1.Ingress{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "networking.k8s.io/v1",
				Kind:       "Ingress",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      IngressName,
				Namespace: IngressNamespace,
				Annotations: map[string]string{
					"example-mutating-admission-webhook": "foo",
				},
			},
			Spec: netv1.IngressSpec{
				Rules: []netv1.IngressRule{
					{Host: "test",
						IngressRuleValue: netv1.IngressRuleValue{HTTP: &netv1.HTTPIngressRuleValue{Paths: []netv1.HTTPIngressPath{
							{
								Path:     "/",
								PathType: &prefix,
								Backend: netv1.IngressBackend{
									Service: &netv1.IngressServiceBackend{
										Name: "test",
										Port: netv1.ServiceBackendPort{
											Number: 80,
										},
									},
								},
							},
						}}}},
				},
			},
		}
		Context("valid", func() {
			It("should create a valid Ingress", func() {
				validateIngressTest(&validIngress, true)
			})
		})

		Context("invalid", func() {
			It("should not create a invalid Ingress that have not annotation key 'example-mutating-admission-webhook'", func() {
				invalidIngress := validIngress.DeepCopy()
				invalidIngress.ObjectMeta.Name = "test-ingress2"
				invalidIngress.ObjectMeta.Annotations = map[string]string{}
				validateIngressTest(&validIngress, false)
			})

			It("should not create a invalid Ingress that have not annotation 'example-mutating-admission-webhook' values 'foo'", func() {
				invalidIngress := validIngress.DeepCopy()
				invalidIngress.ObjectMeta.Name = "test-ingress3"
				invalidIngress.ObjectMeta.Annotations = map[string]string{
					"example-mutating-admission-webhook": "bar",
				}
				validateIngressTest(&validIngress, false)
			})
		})
	})
})
