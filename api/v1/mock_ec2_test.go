package v1

import (
	"context"
	"reflect"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.o-in.dwango.co.jp/naari3/ingress-sg-validator/pkg/aws/services"
)

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
