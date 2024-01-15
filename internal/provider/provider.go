// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure K8sProvider satisfies various provider interfaces.
var _ provider.Provider = &K8sProvider{}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &K8sProvider{
			version: version,
		}
	}
}

// K8sProvider defines the provider implementation.
type K8sProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// Metadata returns the provider type name.
func (p *K8sProvider) Metadata(ctx context.Context, request provider.MetadataRequest, response *provider.MetadataResponse) {
	response.TypeName = "custom-resource"
	response.Version = p.version
}

// Schema defines the provider-level schema for configuration data.
func (p *K8sProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{

		Attributes: map[string]schema.Attribute{
			"host": schema.StringAttribute{
				MarkdownDescription: "The address of the API server host to connect to.",
				Required:            true,
			},
			"token": schema.StringAttribute{
				MarkdownDescription: "The bearer token to use in order to authenticate against the API server; if not specified, a username and password combination should be provided instead.",
				Optional:            true,
				Sensitive:           true,
			},
			"username": schema.StringAttribute{
				MarkdownDescription: "The username to use, in combination with a password, in order to authenticate against the API server. It a bearer token has been provided, it will take precedence over username and password basic authentication",
				Optional:            true,
			},
			"password": schema.StringAttribute{
				MarkdownDescription: "The password to use, in combination with a username, in order to authenticate against the API server. It a bearer token has been provided, it will take precedence over username and password basic authentication",
				Optional:            true,
				Sensitive:           true,
			},
		},
	}
}

// K8sProviderModel describes the provider data model.
type K8sProviderModel struct {
	Host     types.String `tfsdk:"host"`
	Token    types.String `tfsdk:"token"`
	Username types.String `tfsdk:"username"`
	Password types.String `tfsdk:"password"`
}

type K8sProviderConfiguration struct {
	Host     string
	Token    string
	Username string
	Password string
}

// Configure prepares a CRD API client for data sources and resources.
func (p *K8sProvider) Configure(ctx context.Context, request provider.ConfigureRequest, response *provider.ConfigureResponse) {
	var model K8sProviderModel

	tflog.Trace(ctx, "configuring Kubernetes Custom Resource Definition client (kubectl)...")

	response.Diagnostics.Append(request.Config.Get(ctx, &model)...)

	if response.Diagnostics.HasError() {
		return
	}

	if model.Host.IsUnknown() {
		response.Diagnostics.AddAttributeError(
			path.Root("host"),
			"Unknown Kubernetes API Server host",
			"The provider cannot create the Kubernetes API client as there is an unknown configuration value for the API Server endpoint. "+
				"Set the value statically in the configuration, or use the K8S_HOST environment variable.",
		)
	}

	if model.Token.IsUnknown() && (model.Username.IsUnknown() || model.Password.IsUnknown()) {
		response.Diagnostics.AddError(
			"Unknown Kubernetes API Server credentials",
			"The provider cannot create the Kubernetes API client as there are no valid credentials specified. "+
				"Statically set the value of either the token, or the username and password combination, in the configuration, or use the K8S_TOKEN, K8S_USERNAME and K8S_PASSWORD environment variables.",
		)
	}

	if response.Diagnostics.HasError() {
		return
	}

	host := os.Getenv("K8S_HOST")
	token := os.Getenv("K8S_TOKEN")
	username := os.Getenv("K8S_USERNAME")
	password := os.Getenv("K8S_PASSWORD")

	if !model.Host.IsNull() {
		host = model.Host.ValueString()
	}

	if !model.Token.IsNull() {
		token = model.Token.ValueString()
	}

	if !model.Username.IsNull() {
		username = model.Username.ValueString()
	}

	if !model.Password.IsNull() {
		password = model.Password.ValueString()
	}

	if host == "" {
		response.Diagnostics.AddAttributeError(
			path.Root("host"),
			"Missing Kubernetes API Server host",
			"The provider cannot create the Kubernetes API client as there is a missing or empty value for the API Server host. "+
				"Set the endpoint value in the configuration or use the K8S_ENDPOINT environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if token == "" && (username == "" || password == "") {
		response.Diagnostics.AddError(
			"Missing Kubernetes API Server credentials",
			"The provider cannot create the Kubernetes API client as there is a missing or empty value for the API Server credentials. "+
				"Set the token value, or a valid username/password combination in the configuration or use the K8S_TOKEN, "+
				"K8S_USERNAME and K8S_PASSWORD environment variables. "+
				"If either is already set, ensure the values are not empty.",
		)
	}

	if response.Diagnostics.HasError() {
		return
	}

	// Example client configuration for model sources and resources
	configuration := K8sProviderConfiguration{
		Host:     host,
		Token:    token,
		Username: username,
		Password: password,
	}
	response.DataSourceData = &configuration
	response.ResourceData = &configuration
}

// Resources defines the resources implemented in the provider.
func (p *K8sProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewCustomResourceDefinition,
		NewCustomResourceInstance,
	}
}

// DataSources defines the data sources implemented in the provider.
func (p *K8sProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewExampleDataSource,
	}
}
