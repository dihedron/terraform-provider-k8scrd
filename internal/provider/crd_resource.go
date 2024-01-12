// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os/exec"
	"text/template"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var (
	_ resource.Resource = &CRDInstanceResource{}
	//_ resource.ResourceWithImportState = &CRDInstanceResource{} // TODO: do we import a state?
	_ resource.ResourceWithConfigure = &CRDInstanceResource{} // TODO: do we reconfigure?
)

func NewCRDInstanceResource() resource.Resource {
	return &CRDInstanceResource{}
}

// CRDInstanceResource defines the resource implementation.
type CRDInstanceResource struct {
	configuration *K8sProviderConfiguration
}

// CRDInstanceResourceModel describes the resource data model.
type CRDInstanceResourceModel struct {
	Attributes types.Map    `tfsdk:"attributes"`
	Template   types.String `tfsdk:"template"`
	Applied    types.String `tfsdk:"applied"`
	Id         types.String `tfsdk:"id"`
}

func (r *CRDInstanceResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_instance"
}

func (r *CRDInstanceResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Resource instance created according to a registered Custom Resource Definition (CRD)",

		Attributes: map[string]schema.Attribute{
			"attributes": schema.MapAttribute{
				MarkdownDescription: "Template attributes",
				Optional:            true,
				ElementType:         types.StringType,
			},
			"template": schema.StringAttribute{
				MarkdownDescription: "The template to use to define the custom resource",
				Required:            true,
			},
			"applied": schema.StringAttribute{
				MarkdownDescription: "The actual YAML used to create the CRD(s)",
				Computed:            true,
			},
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Identifier for the Custom Resource Definition",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *CRDInstanceResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	configuration, ok := req.ProviderData.(*K8sProviderConfiguration)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *K8sProviderConfiguration, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.configuration = configuration
}

// Create creates the resource by running kubectl apply.
func (r *CRDInstanceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {

	var model CRDInstanceResourceModel

	// read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// create the resource
	r.createOrUpdate(ctx, &model, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

// Read does nothing really.
func (r *CRDInstanceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data CRDInstanceResourceModel

	// read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update re-runs kubectl apply with the whole file and relies on Kubernetes to perform all the necessary diffs.
func (r *CRDInstanceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var model CRDInstanceResourceModel

	// read Terraform plan model into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// update the resource
	r.createOrUpdate(ctx, &model, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *CRDInstanceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data CRDInstanceResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// If applicable, this is a great opportunity to initialize any necessary
	// provider client data and make a call using it.
	// httpResp, err := r.client.Do(httpReq)
	// if err != nil {
	//     resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete example, got error: %s", err))
	//     return
	// }
}

func (r *CRDInstanceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *CRDInstanceResource) createOrUpdate(ctx context.Context, model *CRDInstanceResourceModel, diagnostics *diag.Diagnostics) {

	tflog.Debug(ctx, "creating new CRD resource")

	tmpl, err := template.New("crd").Parse(model.Template.ValueString())
	if err != nil {
		//return err
		diagnostics.AddError(
			"Invalid custom Resource Definition template.",
			fmt.Sprintf("The provided Custom resource Definition template is not valid as it could not be parsed: %v", err),
		)
		return
	}

	attributes := map[string]string{}
	if !model.Template.IsNull() {
		a, diags := model.Attributes.ToMapValue(ctx)
		diagnostics.Append(diags...)
		if diagnostics.HasError() {
			return
		}

		for key, value := range a.Elements() {
			tflog.Debug(ctx, fmt.Sprintf("adding element %s => %s", key, value.String()))
			attributes[key] = value.String()
		}
	}

	var buffer bytes.Buffer
	if err = tmpl.Execute(&buffer, attributes); err != nil {
		diagnostics.AddError(
			"Invalid custom Resource Definition template.",
			fmt.Sprintf("The provided Custom Resource Definition template is not valid as it could not be parsed: %v", err),
		)
		return
	}

	tflog.Debug(ctx, fmt.Sprintf("template after applying values: %s", buffer.String()))

	var cmd *exec.Cmd
	if r.configuration.Token != "" {
		cmd = exec.Command("/data/workspaces/gomods/terraform-provider-k8scrd/examples/kubectl/kubectl", "apply", "--server", r.configuration.Host, "--token", r.configuration.Token, "--output", "json")
	} else if r.configuration.Username != "" && r.configuration.Password != "" {
		cmd = exec.Command("/data/workspaces/gomods/terraform-provider-k8scrd/examples/kubectl/kubectl", "apply", "--server", r.configuration.Host, "--username", r.configuration.Username, "--password", r.configuration.Password, "--output", "json")
	}
	stdin, err := cmd.StdinPipe()
	if err != nil {
		diagnostics.AddError(
			"Error executing external kubectl command.",
			fmt.Sprintf("The kubectl apply command could not be executed: %v", err),
		)
		return
	}

	go func() {
		defer stdin.Close()
		io.WriteString(stdin, buffer.String())
	}()

	out, err := cmd.CombinedOutput()
	if err != nil {
		diagnostics.AddError(
			"Error reading kubectl command output.",
			fmt.Sprintf("The kubectl apply command's output could not be read: %v", err),
		)
		return
	}

	tflog.Debug(ctx, fmt.Sprintf("kubectl apply's output: %s", out))

	// For the purposes of this example code, hardcoding a response value to
	// save into the Terraform state.
	model.Id = types.StringValue("example-id") // TODO: fixme
	model.Applied = types.StringValue(buffer.String())

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Trace(ctx, "created or updated the resource")
}
