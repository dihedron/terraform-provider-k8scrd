// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// Ensure provider defined types fully satisfy framework interfaces.
var (
	_ resource.Resource = &CustomResourceDefinition{}
	//_ resource.ResourceWithImportState = &CustomResourceDefinition{} // TODO: do we import a state?
	_ resource.ResourceWithConfigure = &CustomResourceDefinition{} // TODO: do we reconfigure?
)

func NewCustomResourceDefinition() resource.Resource {
	return &CustomResourceDefinition{}
}

// CustomResourceDefinition defines the resource implementation.
type CustomResourceDefinition struct {
	KubectlResource
}

// CustomResourceDefinitionModel describes the resource data model.
type CustomResourceDefinitionModel struct {
	KubectlResourceModel
}

func (r *CustomResourceDefinition) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_definition"
}
