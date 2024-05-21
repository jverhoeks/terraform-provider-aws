// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_ec2_transit_gateway_multicast_domain")
func DataSourceTransitGatewayMulticastDomain() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceTransitGatewayMulticastDomainRead,

		Timeouts: &schema.ResourceTimeout{
			Read: schema.DefaultTimeout(20 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"associations": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrSubnetID: {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrTransitGatewayAttachmentID: {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"auto_accept_shared_associations": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrFilter: customFiltersSchema(),
			"igmpv2_support": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"members": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"group_ip_address": {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrNetworkInterfaceID: {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			names.AttrOwnerID: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"sources": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"group_ip_address": {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrNetworkInterfaceID: {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			names.AttrState: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"static_sources_support": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags: tftags.TagsSchemaComputed(),
			names.AttrTransitGatewayAttachmentID: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTransitGatewayID: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"transit_gateway_multicast_domain_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
		},
	}
}

func dataSourceTransitGatewayMulticastDomainRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).EC2Conn(ctx)
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	input := &ec2.DescribeTransitGatewayMulticastDomainsInput{}

	if v, ok := d.GetOk("transit_gateway_multicast_domain_id"); ok {
		input.TransitGatewayMulticastDomainIds = aws.StringSlice([]string{v.(string)})
	}

	input.Filters = append(input.Filters, newCustomFilterList(
		d.Get(names.AttrFilter).(*schema.Set),
	)...)

	if len(input.Filters) == 0 {
		input.Filters = nil
	}

	transitGatewayMulticastDomain, err := FindTransitGatewayMulticastDomain(ctx, conn, input)

	if err != nil {
		return sdkdiag.AppendFromErr(diags, tfresource.SingularDataSourceFindError("EC2 Transit Gateway Multicast Domain", err))
	}

	d.SetId(aws.StringValue(transitGatewayMulticastDomain.TransitGatewayMulticastDomainId))
	d.Set(names.AttrARN, transitGatewayMulticastDomain.TransitGatewayMulticastDomainArn)
	d.Set("auto_accept_shared_associations", transitGatewayMulticastDomain.Options.AutoAcceptSharedAssociations)
	d.Set("igmpv2_support", transitGatewayMulticastDomain.Options.Igmpv2Support)
	d.Set(names.AttrOwnerID, transitGatewayMulticastDomain.OwnerId)
	d.Set(names.AttrState, transitGatewayMulticastDomain.State)
	d.Set("static_sources_support", transitGatewayMulticastDomain.Options.StaticSourcesSupport)
	d.Set(names.AttrTransitGatewayID, transitGatewayMulticastDomain.TransitGatewayId)
	d.Set("transit_gateway_multicast_domain_id", transitGatewayMulticastDomain.TransitGatewayMulticastDomainId)

	if err := d.Set(names.AttrTags, KeyValueTags(ctx, transitGatewayMulticastDomain.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags: %s", err)
	}

	associations, err := FindTransitGatewayMulticastDomainAssociations(ctx, conn, &ec2.GetTransitGatewayMulticastDomainAssociationsInput{
		TransitGatewayMulticastDomainId: aws.String(d.Id()),
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "listing EC2 Transit Gateway Multicast Domain Associations (%s): %s", d.Id(), err)
	}

	if err := d.Set("associations", flattenTransitGatewayMulticastDomainAssociations(associations)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting associations: %s", err)
	}

	members, err := FindTransitGatewayMulticastGroups(ctx, conn, &ec2.SearchTransitGatewayMulticastGroupsInput{
		Filters: newAttributeFilterList(map[string]string{
			"is-group-member": "true",
			"is-group-source": "false",
		}),
		TransitGatewayMulticastDomainId: aws.String(d.Id()),
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "listing EC2 Transit Gateway Multicast Group Members (%s): %s", d.Id(), err)
	}

	if err := d.Set("members", flattenTransitGatewayMulticastGroups(members)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting members: %s", err)
	}

	sources, err := FindTransitGatewayMulticastGroups(ctx, conn, &ec2.SearchTransitGatewayMulticastGroupsInput{
		Filters: newAttributeFilterList(map[string]string{
			"is-group-member": "false",
			"is-group-source": "true",
		}),
		TransitGatewayMulticastDomainId: aws.String(d.Id()),
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "listing EC2 Transit Gateway Multicast Group Members (%s): %s", d.Id(), err)
	}

	if err := d.Set("sources", flattenTransitGatewayMulticastGroups(sources)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting sources: %s", err)
	}

	return diags
}

func flattenTransitGatewayMulticastDomainAssociation(apiObject *ec2.TransitGatewayMulticastDomainAssociation) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.Subnet.SubnetId; v != nil {
		tfMap[names.AttrSubnetID] = aws.StringValue(v)
	}

	if v := apiObject.TransitGatewayAttachmentId; v != nil {
		tfMap[names.AttrTransitGatewayAttachmentID] = aws.StringValue(v)
	}

	return tfMap
}

func flattenTransitGatewayMulticastDomainAssociations(apiObjects []*ec2.TransitGatewayMulticastDomainAssociation) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		if apiObject == nil {
			continue
		}

		tfList = append(tfList, flattenTransitGatewayMulticastDomainAssociation(apiObject))
	}

	return tfList
}

func flattenTransitGatewayMulticastGroup(apiObject *ec2.TransitGatewayMulticastGroup) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.GroupIpAddress; v != nil {
		tfMap["group_ip_address"] = aws.StringValue(v)
	}

	if v := apiObject.NetworkInterfaceId; v != nil {
		tfMap[names.AttrNetworkInterfaceID] = aws.StringValue(v)
	}

	return tfMap
}

func flattenTransitGatewayMulticastGroups(apiObjects []*ec2.TransitGatewayMulticastGroup) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		if apiObject == nil {
			continue
		}

		tfList = append(tfList, flattenTransitGatewayMulticastGroup(apiObject))
	}

	return tfList
}
