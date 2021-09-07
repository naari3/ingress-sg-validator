package services

import (
	"context"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
)

type EC2 interface {
	ec2iface.EC2API

	DescribeSecurityGroupsAsList(ctx context.Context, input *ec2.DescribeSecurityGroupsInput) ([]*ec2.SecurityGroup, error)
}

type defaultEC2 struct {
	ec2iface.EC2API
}

func NewEC2(session *session.Session) EC2 {
	return &defaultEC2{
		EC2API: ec2.New(session),
	}
}

func (c *defaultEC2) DescribeSecurityGroupsAsList(ctx context.Context, input *ec2.DescribeSecurityGroupsInput) ([]*ec2.SecurityGroup, error) {
	var result []*ec2.SecurityGroup
	if err := c.DescribeSecurityGroupsPagesWithContext(ctx, input, func(output *ec2.DescribeSecurityGroupsOutput, _ bool) bool {
		result = append(result, output.SecurityGroups...)
		return true
	}); err != nil {
		return nil, err
	}
	return result, nil
}
