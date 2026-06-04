// Copyright IBM Corp. 2017, 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"crypto/rand"
	"math/big"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = (*failureResource)(nil)

func NewFailureResource() resource.Resource {
	return &failureResource{}
}

type failureResource struct{}

func (r *failureResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_failure"
}

func (r *failureResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "The resource `random_failure` intentionally fails lifecycle operations based on configured percentages.",
		Attributes: map[string]schema.Attribute{
			"create_failure_percentage": schema.Int64Attribute{
				Description: "Percentage chance that the create operation fails. Must be between `0` and `100`, inclusive. Default value is `0`.",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(0),
				Validators: []validator.Int64{
					int64validator.AtLeast(0),
					int64validator.AtMost(100),
				},
			},
			"update_failure_percentage": schema.Int64Attribute{
				Description: "Percentage chance that the update operation fails. Must be between `0` and `100`, inclusive. Default value is `0`.",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(0),
				Validators: []validator.Int64{
					int64validator.AtLeast(0),
					int64validator.AtMost(100),
				},
			},
			"destroy_failure_percentage": schema.Int64Attribute{
				Description: "Percentage chance that the destroy operation fails. Must be between `0` and `100`, inclusive. Default value is `0`.",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(0),
				Validators: []validator.Int64{
					int64validator.AtLeast(0),
					int64validator.AtMost(100),
				},
			},
			"id": schema.StringAttribute{
				Description: "A static value used internally by Terraform, this should not be referenced in configurations.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *failureResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan failureModelV0

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	shouldFail, err := shouldFailByPercentage(plan.CreateFailurePercentage.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError(
			"Create random_failure failed",
			"Unable to generate random value for create failure simulation: "+err.Error(),
		)
		return
	}

	if shouldFail {
		resp.Diagnostics.AddError(
			"Create random_failure failed",
			"create operation failed by random_failure simulation",
		)
		return
	}

	plan.ID = types.StringValue("-")

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Read does not need to perform any operations as the state in ReadResourceResponse is already populated.
func (r *failureResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
}

func (r *failureResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan failureModelV0

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	shouldFail, err := shouldFailByPercentage(plan.UpdateFailurePercentage.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError(
			"Update random_failure failed",
			"Unable to generate random value for update failure simulation: "+err.Error(),
		)
		return
	}

	if shouldFail {
		resp.Diagnostics.AddError(
			"Update random_failure failed",
			"update operation failed by random_failure simulation",
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete does not need to explicitly call resp.State.RemoveResource() as this is automatically handled by the
// [framework](https://github.com/hashicorp/terraform-plugin-framework/pull/301).
func (r *failureResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state failureModelV0

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	shouldFail, err := shouldFailByPercentage(state.DestroyFailurePercentage.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError(
			"Delete random_failure failed",
			"Unable to generate random value for destroy failure simulation: "+err.Error(),
		)
		return
	}

	if shouldFail {
		resp.Diagnostics.AddError(
			"Delete random_failure failed",
			"destroy operation failed by random_failure simulation",
		)
		return
	}
}

func shouldFailByPercentage(percentage int64) (bool, error) {
	if percentage <= 0 {
		return false, nil
	}

	if percentage >= 100 {
		return true, nil
	}

	randomValue, err := rand.Int(rand.Reader, big.NewInt(100))
	if err != nil {
		return false, err
	}

	return randomValue.Int64() < percentage, nil
}

type failureModelV0 struct {
	ID                       types.String `tfsdk:"id"`
	CreateFailurePercentage  types.Int64  `tfsdk:"create_failure_percentage"`
	UpdateFailurePercentage  types.Int64  `tfsdk:"update_failure_percentage"`
	DestroyFailurePercentage types.Int64  `tfsdk:"destroy_failure_percentage"`
}
