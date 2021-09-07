package pkg

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.o-in.dwango.co.jp/naari3/ingress-sg-validator/pkg/aws/services"
	"sigs.k8s.io/controller-runtime/pkg/envtest/printer"

	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestAPIs(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecsWithDefaultAndCustomReporters(t,
		"SG Validator",
		[]Reporter{printer.NewlineReporter{}})
}

type mockEC2 struct {
	services.EC2
	store []*ec2.SecurityGroup
}

// Only support 'GroupIds' and part of 'Filter'
func (c *mockEC2) DescribeSecurityGroupsAsList(ctx context.Context, input *ec2.DescribeSecurityGroupsInput) ([]*ec2.SecurityGroup, error) {
	var result []*ec2.SecurityGroup

	groupIds := []string{}
	for _, gid := range input.GroupIds {
		groupIds = append(groupIds, *gid)
	}

	var nameFilter *ec2.Filter
	for _, f := range input.Filters {
		if *f.Name == "tag:Name" {
			nameFilter = f
			break
		}
	}

	names := []string{}
	if nameFilter != nil {
		for _, v := range nameFilter.Values {
			names = append(names, *v)
		}
	}

	for _, sg := range c.store {
		if sg != nil && contains(groupIds, *sg.GroupId) {
			result = append(result, sg)
		} else if nameFilter != nil && *nameFilter.Name == "tag:Name" && contains(names, *(sg.GroupName)) {
			result = append(result, sg)
		}
	}
	return result, nil
}

func contains(list interface{}, elem interface{}) bool {
	listV := reflect.ValueOf(list)

	if listV.Kind() == reflect.Slice {
		for i := 0; i < listV.Len(); i++ {
			item := listV.Index(i).Interface()
			if !reflect.TypeOf(elem).ConvertibleTo(reflect.TypeOf(item)) {
				continue
			}
			target := reflect.ValueOf(elem).Convert(reflect.TypeOf(item)).Interface()
			if ok := reflect.DeepEqual(item, target); ok {
				return true
			}
		}
	}
	return false
}

var _ = Describe("Ingress Webhook", func() {
	var (
		ing networkingv1.Ingress
		v   SGValidator
	)
	BeforeEach(func() {
		validSGIDs := []string{"sg-1", "sg-2", "sg-3"}
		validSGNames := []string{"foo1", "foo2", "foo3"}
		sgs := make([]*ec2.SecurityGroup, 0)

		for _, _sgID := range validSGIDs {
			sgID := _sgID
			sg := ec2.SecurityGroup{
				Description: &sgID,
				GroupId:     &sgID,
				GroupName:   &sgID,
			}
			sgs = append(sgs, &sg)
		}
		nameIDCount := 100
		for _, _sgName := range validSGNames {
			sgName := _sgName
			ID := fmt.Sprintf("sg-%d", nameIDCount)
			sg := ec2.SecurityGroup{
				Description: &sgName,
				GroupId:     &ID,
				GroupName:   &sgName,
			}
			sgs = append(sgs, &sg)
			nameIDCount++
		}

		v = NewValidator(&mockEC2{
			store: sgs,
		})
	})

	BeforeEach(func() {
		ing = networkingv1.Ingress{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "networking.k8s.io/v1",
				Kind:       "Ingress",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:        "ingress",
				Annotations: map[string]string{},
			},
		}
	})

	Context("validating", func() {
		Context("valid", func() {
			It("should valid empty annotations", func() {
				ctx := context.Background()

				ing.ObjectMeta.Annotations = make(map[string]string)

				err := v.ValidateAnnotation(ctx, &ing)
				Expect(err).NotTo(HaveOccurred(), "Ingress: %v", ing)
			})

			It("should valid exist sg annotations", func() {
				ctx := context.Background()

				ing.ObjectMeta.Annotations = make(map[string]string)
				ing.ObjectMeta.Annotations[sgKey] = "sg-1"

				err := v.ValidateAnnotation(ctx, &ing)
				Expect(err).NotTo(HaveOccurred(), "Ingress: %v", ing)
			})

			It("should valid multiple exist sg annotations", func() {
				ctx := context.Background()

				ing.ObjectMeta.Annotations = make(map[string]string)
				ing.ObjectMeta.Annotations[sgKey] = "sg-1, sg-2, sg-3, foo1, foo2, foo3"

				err := v.ValidateAnnotation(ctx, &ing)
				Expect(err).NotTo(HaveOccurred(), "Ingress: %v", ing)
			})
		})

		Context("invalid", func() {
			It("should invalid non-exist sg annotations", func() {
				ctx := context.Background()

				ing.ObjectMeta.Annotations = make(map[string]string)
				ing.ObjectMeta.Annotations[sgKey] = "sg-10"

				err := v.ValidateAnnotation(ctx, &ing)
				Expect(err).To(HaveOccurred(), "Ingress: %v", ing)
			})

			It("should invalid non-exist and exist sg annotations", func() {
				ctx := context.Background()

				ing.ObjectMeta.Annotations = make(map[string]string)
				ing.ObjectMeta.Annotations[sgKey] = "sg-1, sg-2, sg-4"

				err := v.ValidateAnnotation(ctx, &ing)
				Expect(err).To(HaveOccurred(), "Ingress: %v", ing)
			})
		})
	})
})
