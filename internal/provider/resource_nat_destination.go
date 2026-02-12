package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ resource.Resource = &NatDestinationResource{}
var _ resource.ResourceWithImportState = &NatDestinationResource{}

func NewNatDestinationResource() resource.Resource {
	return &NatDestinationResource{}
}

type NatDestinationResource struct {
	client *Client
}

type NatDestinationResourceModel struct {
	ID              types.String `tfsdk:"id"`
	Enabled         types.Bool   `tfsdk:"enabled"`
	Sequence        types.Int64  `tfsdk:"sequence"`
	Interface       types.String `tfsdk:"interface"`
	Protocol        types.String `tfsdk:"protocol"`
	IPProtocol      types.String `tfsdk:"ip_protocol"`
	SourceNet       types.String `tfsdk:"source_net"`
	SourcePort      types.String `tfsdk:"source_port"`
	SourceNot       types.Bool   `tfsdk:"source_not"`
	DestinationNet  types.String `tfsdk:"destination_net"`
	DestinationPort types.String `tfsdk:"destination_port"`
	DestinationNot  types.Bool   `tfsdk:"destination_not"`
	TargetIP        types.String `tfsdk:"target_ip"`
	TargetPort      types.String `tfsdk:"target_port"`
	Description     types.String `tfsdk:"description"`
	Log             types.Bool   `tfsdk:"log"`
	NATReflection   types.String `tfsdk:"nat_reflection"`
}

func (r *NatDestinationResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_nat_destination"
}

func (r *NatDestinationResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages Destination NAT (Port Forward) rules in OPNsense",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "NAT rule UUID",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"enabled": schema.BoolAttribute{
				MarkdownDescription: "Enable this NAT rule (default: true)",
				Optional:            true,
				Computed:            true,
			},
			"sequence": schema.Int64Attribute{
				MarkdownDescription: "Rule sequence/priority (lower = higher priority)",
				Optional:            true,
			},
			"interface": schema.StringAttribute{
				MarkdownDescription: "Interface (e.g., 'wan')",
				Required:            true,
			},
			"protocol": schema.StringAttribute{
				MarkdownDescription: "Protocol: 'tcp', 'udp', or 'tcp/udp'",
				Required:            true,
			},
			"ip_protocol": schema.StringAttribute{
				MarkdownDescription: "IP protocol: 'inet' (IPv4), 'inet6' (IPv6), or 'inet46' (both)",
				Optional:            true,
			},
			"source_net": schema.StringAttribute{
				MarkdownDescription: "Source network/address",
				Optional:            true,
			},
			"source_port": schema.StringAttribute{
				MarkdownDescription: "Source port",
				Optional:            true,
			},
			"source_not": schema.BoolAttribute{
				MarkdownDescription: "Invert source match",
				Optional:            true,
			},
			"destination_net": schema.StringAttribute{
				MarkdownDescription: "Destination network/address (e.g., 'wanip', 'any')",
				Optional:            true,
			},
			"destination_port": schema.StringAttribute{
				MarkdownDescription: "External/destination port (required)",
				Required:            true,
			},
			"destination_not": schema.BoolAttribute{
				MarkdownDescription: "Invert destination match",
				Optional:            true,
			},
			"target_ip": schema.StringAttribute{
				MarkdownDescription: "Internal target IP address or alias (required)",
				Required:            true,
			},
			"target_port": schema.StringAttribute{
				MarkdownDescription: "Internal target port (required)",
				Required:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Description for this NAT rule",
				Optional:            true,
			},
			"log": schema.BoolAttribute{
				MarkdownDescription: "Log packets matching this rule",
				Optional:            true,
			},
			"nat_reflection": schema.StringAttribute{
				MarkdownDescription: "NAT reflection: 'enable', 'purenat', 'disable'",
				Optional:            true,
			},
		},
	}
}

func (r *NatDestinationResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *Client, got: %T.", req.ProviderData),
		)
		return
	}

	r.client = client
}

func (r *NatDestinationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data NatDestinationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Build NAT destination rule payload
	// ACTUAL field names from API (using dots and hyphens!):
	natData := map[string]interface{}{
		"destination": map[string]interface{}{
			"interface":          data.Interface.ValueString(),
			"protocol":           data.Protocol.ValueString(),
			"destination.port":   data.DestinationPort.ValueString(), // Uses dot!
			"target":             data.TargetIP.ValueString(),
			"local-port":         data.TargetPort.ValueString(), // Uses hyphen!
		},
	}

	// Disabled field (0 = enabled, 1 = disabled - inverted!)
	if !data.Enabled.IsNull() {
		if data.Enabled.ValueBool() {
			natData["destination"].(map[string]interface{})["disabled"] = "0"
		} else {
			natData["destination"].(map[string]interface{})["disabled"] = "1"
		}
	} else {
		natData["destination"].(map[string]interface{})["disabled"] = "0"
	}

	// Sequence
	if !data.Sequence.IsNull() {
		natData["destination"].(map[string]interface{})["sequence"] = fmt.Sprintf("%d", data.Sequence.ValueInt64())
	}

	// IP Protocol
	if !data.IPProtocol.IsNull() {
		natData["destination"].(map[string]interface{})["ipprotocol"] = data.IPProtocol.ValueString()
	}

	// Optional source fields
	if !data.SourceNet.IsNull() {
		natData["destination"].(map[string]interface{})["source.network"] = data.SourceNet.ValueString() // Uses dot!
	}

	if !data.SourcePort.IsNull() {
		natData["destination"].(map[string]interface{})["source.port"] = data.SourcePort.ValueString() // Uses dot!
	}

	if !data.SourceNot.IsNull() && data.SourceNot.ValueBool() {
		natData["destination"].(map[string]interface{})["source.not"] = "1" // Uses dot!
	}

	// Optional destination fields
	if !data.DestinationNet.IsNull() {
		natData["destination"].(map[string]interface{})["destination.network"] = data.DestinationNet.ValueString() // Uses dot!
	}

	if !data.DestinationNot.IsNull() && data.DestinationNot.ValueBool() {
		natData["destination"].(map[string]interface{})["destination.not"] = "1" // Uses dot!
	}

	// Description
	if !data.Description.IsNull() {
		natData["destination"].(map[string]interface{})["descr"] = data.Description.ValueString()
	}

	// Log
	if !data.Log.IsNull() && data.Log.ValueBool() {
		natData["destination"].(map[string]interface{})["log"] = "1"
	}

	// NAT Reflection
	if !data.NATReflection.IsNull() {
		natData["destination"].(map[string]interface{})["natreflection"] = data.NATReflection.ValueString()
	}

	jsonData, _ := json.Marshal(natData)
	tflog.Debug(ctx, "Creating NAT destination rule", map[string]any{"payload": string(jsonData)})

	// API endpoint
	url := fmt.Sprintf("%s/api/firewall/d_nat/add_rule", r.client.Host)
	httpReq, _ := http.NewRequestWithContext(ctx, "POST", url, strings.NewReader(string(jsonData)))
	httpReq.SetBasicAuth(r.client.ApiKey, r.client.ApiSecret)
	httpReq.Header.Set("Content-Type", "application/json")

	httpResp, err := r.client.client.Do(httpReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create NAT rule: %s", err))
		return
	}
	defer httpResp.Body.Close()

	body, _ := io.ReadAll(httpResp.Body)
	tflog.Debug(ctx, "NAT rule created", map[string]any{"response": string(body)})

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to parse response: %s", err))
		return
	}

	// Check for failed result
	if resultStr, ok := result["result"].(string); ok && resultStr == "failed" {
		resp.Diagnostics.AddError("API Error", fmt.Sprintf("API returned failed: %s", string(body)))
		return
	}

	// Extract UUID
	if uuid, ok := result["uuid"].(string); ok {
		data.ID = types.StringValue(uuid)
	} else if resultStr, ok := result["result"].(string); ok {
		data.ID = types.StringValue(resultStr)
	} else {
		resp.Diagnostics.AddError("API Error", fmt.Sprintf("No UUID returned from API: %s", string(body)))
		return
	}

	// Apply the configuration
	applyURL := fmt.Sprintf("%s/api/firewall/d_nat/apply", r.client.Host)
	applyReq, _ := http.NewRequestWithContext(ctx, "POST", applyURL, nil)
	applyReq.SetBasicAuth(r.client.ApiKey, r.client.ApiSecret)

	applyResp, err := r.client.client.Do(applyReq)
	if err != nil {
		tflog.Warn(ctx, "Failed to apply NAT configuration", map[string]any{"error": err.Error()})
	} else {
		defer applyResp.Body.Close()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *NatDestinationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data NatDestinationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	url := fmt.Sprintf("%s/api/firewall/d_nat/get_rule/%s", r.client.Host, data.ID.ValueString())
	httpReq, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
	httpReq.SetBasicAuth(r.client.ApiKey, r.client.ApiSecret)

	httpResp, err := r.client.client.Do(httpReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read NAT rule: %s", err))
		return
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode == 404 {
		resp.State.RemoveResource(ctx)
		return
	}

	body, _ := io.ReadAll(httpResp.Body)

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to parse response: %s", err))
		return
	}

	// Parse from search_rule format
	if iface, ok := result["interface"].(string); ok {
		data.Interface = types.StringValue(iface)
	}
	if proto, ok := result["protocol"].(string); ok {
		data.Protocol = types.StringValue(proto)
	}
	if target, ok := result["target"].(string); ok {
		data.TargetIP = types.StringValue(target)
	}
	if descr, ok := result["descr"].(string); ok {
		data.Description = types.StringValue(descr)
	}
	if disabled, ok := result["disabled"].(string); ok {
		data.Enabled = types.BoolValue(disabled == "0") // Inverted!
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *NatDestinationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data NatDestinationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	natData := map[string]interface{}{
		"destination": map[string]interface{}{
			"interface":        data.Interface.ValueString(),
			"protocol":         data.Protocol.ValueString(),
			"destination.port": data.DestinationPort.ValueString(),
			"target":           data.TargetIP.ValueString(),
			"local-port":       data.TargetPort.ValueString(),
		},
	}

	if !data.Enabled.IsNull() {
		if data.Enabled.ValueBool() {
			natData["destination"].(map[string]interface{})["disabled"] = "0"
		} else {
			natData["destination"].(map[string]interface{})["disabled"] = "1"
		}
	}

	if !data.Sequence.IsNull() {
		natData["destination"].(map[string]interface{})["sequence"] = fmt.Sprintf("%d", data.Sequence.ValueInt64())
	}

	if !data.IPProtocol.IsNull() {
		natData["destination"].(map[string]interface{})["ipprotocol"] = data.IPProtocol.ValueString()
	}

	if !data.SourceNet.IsNull() {
		natData["destination"].(map[string]interface{})["source.network"] = data.SourceNet.ValueString()
	}

	if !data.SourcePort.IsNull() {
		natData["destination"].(map[string]interface{})["source.port"] = data.SourcePort.ValueString()
	}

	if !data.SourceNot.IsNull() && data.SourceNot.ValueBool() {
		natData["destination"].(map[string]interface{})["source.not"] = "1"
	}

	if !data.DestinationNet.IsNull() {
		natData["destination"].(map[string]interface{})["destination.network"] = data.DestinationNet.ValueString()
	}

	if !data.DestinationNot.IsNull() && data.DestinationNot.ValueBool() {
		natData["destination"].(map[string]interface{})["destination.not"] = "1"
	}

	if !data.Description.IsNull() {
		natData["destination"].(map[string]interface{})["descr"] = data.Description.ValueString()
	}

	if !data.Log.IsNull() && data.Log.ValueBool() {
		natData["destination"].(map[string]interface{})["log"] = "1"
	}

	if !data.NATReflection.IsNull() {
		natData["destination"].(map[string]interface{})["natreflection"] = data.NATReflection.ValueString()
	}

	jsonData, _ := json.Marshal(natData)

	url := fmt.Sprintf("%s/api/firewall/d_nat/set_rule/%s", r.client.Host, data.ID.ValueString())
	httpReq, _ := http.NewRequestWithContext(ctx, "POST", url, strings.NewReader(string(jsonData)))
	httpReq.SetBasicAuth(r.client.ApiKey, r.client.ApiSecret)
	httpReq.Header.Set("Content-Type", "application/json")

	httpResp, err := r.client.client.Do(httpReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update NAT rule: %s", err))
		return
	}
	defer httpResp.Body.Close()

	// Apply
	applyURL := fmt.Sprintf("%s/api/firewall/d_nat/apply", r.client.Host)
	applyReq, _ := http.NewRequestWithContext(ctx, "POST", applyURL, nil)
	applyReq.SetBasicAuth(r.client.ApiKey, r.client.ApiSecret)

	applyResp, err := r.client.client.Do(applyReq)
	if err != nil {
		tflog.Warn(ctx, "Failed to apply NAT configuration", map[string]any{"error": err.Error()})
	} else {
		defer applyResp.Body.Close()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *NatDestinationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data NatDestinationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	url := fmt.Sprintf("%s/api/firewall/d_nat/del_rule/%s", r.client.Host, data.ID.ValueString())
	httpReq, _ := http.NewRequestWithContext(ctx, "POST", url, nil)
	httpReq.SetBasicAuth(r.client.ApiKey, r.client.ApiSecret)

	httpResp, err := r.client.client.Do(httpReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete NAT rule: %s", err))
		return
	}
	defer httpResp.Body.Close()

	// Apply
	applyURL := fmt.Sprintf("%s/api/firewall/d_nat/apply", r.client.Host)
	applyReq, _ := http.NewRequestWithContext(ctx, "POST", applyURL, nil)
	applyReq.SetBasicAuth(r.client.ApiKey, r.client.ApiSecret)

	applyResp, err := r.client.client.Do(applyReq)
	if err != nil {
		tflog.Warn(ctx, "Failed to apply NAT configuration", map[string]any{"error": err.Error()})
	} else {
		defer applyResp.Body.Close()
	}
}

func (r *NatDestinationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}