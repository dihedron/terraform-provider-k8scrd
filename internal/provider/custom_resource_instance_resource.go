// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// Ensure provider defined types fully satisfy framework interfaces.
var (
	_ resource.Resource = &CustomResourceInstance{}
	//_ resource.ResourceWithImportState = &CustomResourceDefinition{} // TODO: do we import a state?
	_ resource.ResourceWithConfigure = &CustomResourceInstance{} // TODO: do we reconfigure?
)

func NewCustomResourceInstance() resource.Resource {
	return &CustomResourceInstance{}
}

// CustomResourceInstance defines the resource implementation.
type CustomResourceInstance struct {
	KubectlResource
}

// CustomResourceInstanceModel describes the resource data model.
type CustomResourceInstanceModel struct {
	KubectlResourceModel
}

func (r *CustomResourceInstance) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_instance"
}
