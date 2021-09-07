package pkg

import (
	"context"
	"strings"

	awssdk "github.com/aws/aws-sdk-go/aws"
	ec2sdk "github.com/aws/aws-sdk-go/service/ec2"
	"github.com/pkg/errors"
	"github.o-in.dwango.co.jp/naari3/ingress-sg-validator/pkg/aws/services"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Ref https://kubernetes-sigs.github.io/aws-load-balancer-controller/guide/ingress/annotations/#group.name
const sgKey = "alb.ingress.kubernetes.io/security-groups"

type SGValidator struct {
	ec2 services.EC2
}

func NewValidator(ec2 services.EC2) SGValidator {
	return SGValidator{
		ec2: ec2,
	}
}

// ValidateAnnotation checks a security group list in annotation are exists
func (v *SGValidator) ValidateAnnotation(ctx context.Context, ing metav1.Object) error {
	rawSGNameOrIDs, found := ing.GetAnnotations()[sgKey]

	if !found {
		return nil
	}

	if _, err := v.resolveSecurityGroupIDsViaNameOrIDSlice(ctx, splitCommaSeparatedString(rawSGNameOrIDs)); err != nil {
		return err
	}

	return nil
}

func (v *SGValidator) resolveSecurityGroupIDsViaNameOrIDSlice(ctx context.Context, sgNameOrIDs []string) ([]string, error) {
	var sgIDs []string
	var sgNames []string
	for _, nameOrID := range sgNameOrIDs {
		if strings.HasPrefix(nameOrID, "sg-") {
			sgIDs = append(sgIDs, nameOrID)
		} else {
			sgNames = append(sgNames, nameOrID)
		}
	}
	var resolvedSGs []*ec2sdk.SecurityGroup
	if len(sgIDs) > 0 {
		req := &ec2sdk.DescribeSecurityGroupsInput{
			GroupIds: awssdk.StringSlice(sgIDs),
		}
		sgs, err := v.ec2.DescribeSecurityGroupsAsList(ctx, req)
		if err != nil {
			return nil, err
		}
		resolvedSGs = append(resolvedSGs, sgs...)
	}
	if len(sgNames) > 0 {
		req := &ec2sdk.DescribeSecurityGroupsInput{
			Filters: []*ec2sdk.Filter{
				{
					Name:   awssdk.String("tag:Name"),
					Values: awssdk.StringSlice(sgNames),
				},
				// TODO?
				// {
				// 	Name:   awssdk.String("vpc-id"),
				// 	Values: awssdk.StringSlice([]string{t.vpcID}),
				// },
			},
		}
		sgs, err := v.ec2.DescribeSecurityGroupsAsList(ctx, req)
		if err != nil {
			return nil, err
		}
		resolvedSGs = append(resolvedSGs, sgs...)
	}
	resolvedSGIDs := make([]string, 0, len(resolvedSGs))
	for _, sg := range resolvedSGs {
		resolvedSGIDs = append(resolvedSGIDs, awssdk.StringValue(sg.GroupId))
	}
	if len(resolvedSGIDs) != len(sgNameOrIDs) {
		return nil, errors.Errorf("couldn't found all securityGroups, nameOrIDs: %v, found: %v", sgNameOrIDs, resolvedSGIDs)
	}
	return resolvedSGIDs, nil
}

func splitCommaSeparatedString(commaSeparatedString string) []string {
	var result []string
	parts := strings.Split(commaSeparatedString, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if len(part) == 0 {
			continue
		}
		result = append(result, part)
	}
	return result
}
