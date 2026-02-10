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

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &FirewallRuleResource{}
var _ resource.ResourceWithImportState = &FirewallRuleResource{}

func NewFirewallRuleResource() resource.Resource {
	return &FirewallRuleResource{}
}

// FirewallRuleResource defines the resource implementation.
type FirewallRuleResource struct {
	client *Client
}

// FirewallRuleResourceModel describes the resource data model.
type FirewallRuleResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Description types.String `tfsdk:"description"`
	Sequence    types.Int64  `tfsdk:"sequence"`
	Interface   types.String `tfsdk:"interface"`
	Direction   types.String `tfsdk:"direction"`
	IPProtocol  types.String `tfsdk:"ip_protocol"`
	Protocol    types.String `tfsdk:"protocol"`
	SourceNet   types.String `tfsdk:"source_net"`
	SourcePort  types.String `tfsdk:"source_port"`
	DestNet     types.String `tfsdk:"destination_net"`
	DestPort    types.String `tfsdk:"destination_port"`
	Gateway     types.String `tfsdk:"gateway"`
	Action      types.String `tfsdk:"action"`
	Enabled     types.Bool   `tfsdk:"enabled"`
	Log         types.Bool   `tfsdk:"log"`
	Quick       types.Bool   `tfsdk:"quick"`
	Invert      types.Bool   `tfsdk:"invert"`
	// Deprecated fields for backward compatibility - kept to avoid state errors
	SourceNot      types.Bool `tfsdk:"source_not"`
	DestinationNot types.Bool `tfsdk:"destination_not"`
	Categories     types.List `tfsdk:"categories"`
}

func (r *FirewallRuleResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_firewall_rule"
}

func (r *FirewallRuleResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages OPNsense firewall rules via the API",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Rule UUID",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Description of the firewall rule",
				Required:            true,
			},
			"sequence": schema.Int64Attribute{
				MarkdownDescription: "Rule sequence/sort order (e.g., 800). Lower numbers are processed first.",
				Optional:            true,
				Computed:            true,
			},
			"interface": schema.StringAttribute{
				MarkdownDescription: "Interface name (e.g., 'wan', 'lan', 'opt1')",
				Optional:            true,
			},
			"direction": schema.StringAttribute{
				MarkdownDescription: "Direction of traffic ('in' or 'out'). Default is 'in'",
				Optional:            true,
			},
			"ip_protocol": schema.StringAttribute{
				MarkdownDescription: "IP protocol version ('inet' for IPv4, 'inet6' for IPv6). Default is 'inet'",
				Optional:            true,
			},
			"protocol": schema.StringAttribute{
				MarkdownDescription: "Protocol (tcp, udp, icmp, any, etc.)",
				Required:            true,
			},
			"source_net": schema.StringAttribute{
				MarkdownDescription: "Source network or IP address (e.g., '192.168.1.0/24', 'any')",
				Required:            true,
			},
			"source_port": schema.StringAttribute{
				MarkdownDescription: "Source port or port range",
				Optional:            true,
			},
			"destination_net": schema.StringAttribute{
				MarkdownDescription: "Destination network or IP address",
				Required:            true,
			},
			"destination_port": schema.StringAttribute{
				MarkdownDescription: "Destination port or port range",
				Optional:            true,
			},
			"gateway": schema.StringAttribute{
				MarkdownDescription: "Gateway to use for this rule (e.g., 'WAN_DHCP', 'WAN_GW', gateway name)",
				Optional:            true,
			},
			"action": schema.StringAttribute{
				MarkdownDescription: "Action to take ('pass', 'block', 'reject'). Default is 'pass'",
				Optional:            true,
			},
			"enabled": schema.BoolAttribute{
				MarkdownDescription: "Whether the rule is enabled. Default is true",
				Optional:            true,
			},
			"log": schema.BoolAttribute{
				MarkdownDescription: "Whether to log packets matching this rule",
				Optional:            true,
			},
			"quick": schema.BoolAttribute{
				MarkdownDescription: "Apply action immediately on match",
				Optional:            true,
			},
			"invert": schema.BoolAttribute{
				MarkdownDescription: "Invert the rule match (NOT operation)",
				Optional:            true,
			},
			"source_not": schema.BoolAttribute{
				MarkdownDescription: "Invert source match (NOT source)",
				Optional:            true,
			},
			"destination_not": schema.BoolAttribute{
				MarkdownDescription: "Invert destination match (NOT destination)",
				Optional:            true,
			},
			"categories": schema.ListAttribute{
				MarkdownDescription: "List of category UUIDs for organizing rules",
				Optional:            true,
				ElementType:         types.StringType,
			},
		},
	}
}

func (r *FirewallRuleResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = client
}

func (r *FirewallRuleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data FirewallRuleResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Prepare rule data
	ruleData := map[string]interface{}{
		"rule": map[string]interface{}{
			"description": data.Description.ValueString(),
			"source_net":  data.SourceNet.ValueString(),
			"destination_net": data.DestNet.ValueString(),
			"protocol":    data.Protocol.ValueString(),
		},
	}

	// Add sequence if provided
	if !data.Sequence.IsNull() {
		ruleData["rule"].(map[string]interface{})["sequence"] = fmt.Sprintf("%d", data.Sequence.ValueInt64())
	}

	// Add optional fields
	if !data.Interface.IsNull() {
		ruleData["rule"].(map[string]interface{})["interface"] = data.Interface.ValueString()
	}
	if !data.Direction.IsNull() {
		ruleData["rule"].(map[string]interface{})["direction"] = data.Direction.ValueString()
	} else {
		ruleData["rule"].(map[string]interface{})["direction"] = "in"
	}
	if !data.IPProtocol.IsNull() {
		ruleData["rule"].(map[string]interface{})["ipprotocol"] = data.IPProtocol.ValueString()
	} else {
		ruleData["rule"].(map[string]interface{})["ipprotocol"] = "inet"
	}
	if !data.SourcePort.IsNull() {
		ruleData["rule"].(map[string]interface{})["source_port"] = data.SourcePort.ValueString()
	}
	if !data.DestPort.IsNull() {
		ruleData["rule"].(map[string]interface{})["destination_port"] = data.DestPort.ValueString()
	}
	if !data.Gateway.IsNull() {
		ruleData["rule"].(map[string]interface{})["gateway"] = data.Gateway.ValueString()
	}
	if !data.Action.IsNull() {
		ruleData["rule"].(map[string]interface{})["action"] = data.Action.ValueString()
	} else {
		ruleData["rule"].(map[string]interface{})["action"] = "pass"
	}
	if !data.Enabled.IsNull() {
		if data.Enabled.ValueBool() {
			ruleData["rule"].(map[string]interface{})["enabled"] = "1"
		} else {
			ruleData["rule"].(map[string]interface{})["enabled"] = "0"
		}
	} else {
		ruleData["rule"].(map[string]interface{})["enabled"] = "1"
	}
	if !data.Log.IsNull() {
		if data.Log.ValueBool() {
			ruleData["rule"].(map[string]interface{})["log"] = "1"
		} else {
			ruleData["rule"].(map[string]interface{})["log"] = "0"
		}
	}
	if !data.Quick.IsNull() {
		if data.Quick.ValueBool() {
			ruleData["rule"].(map[string]interface{})["quick"] = "1"
		} else {
			ruleData["rule"].(map[string]interface{})["quick"] = "0"
		}
	}
	if !data.Invert.IsNull() {
		if data.Invert.ValueBool() {
			ruleData["rule"].(map[string]interface{})["destination_not"] = "1"
		} else {
			ruleData["rule"].(map[string]interface{})["destination_not"] = "0"
		}
	}
	// Handle deprecated but still functional destination_not field
	if !data.DestinationNot.IsNull() {
		if data.DestinationNot.ValueBool() {
			ruleData["rule"].(map[string]interface{})["destination_not"] = "1"
		} else {
			ruleData["rule"].(map[string]interface{})["destination_not"] = "0"
		}
	}
	// Handle deprecated but still functional source_not field
	if !data.SourceNot.IsNull() {
		if data.SourceNot.ValueBool() {
			ruleData["rule"].(map[string]interface{})["source_not"] = "1"
		} else {
			ruleData["rule"].(map[string]interface{})["source_not"] = "0"
		}
	}
	if !data.Categories.IsNull() && !data.Categories.IsUnknown() {
		tflog.Debug(ctx, "Categories field is present", map[string]any{
			"isNull": data.Categories.IsNull(),
			"isUnknown": data.Categories.IsUnknown(),
		})
		var categories []string
		diags := data.Categories.ElementsAs(ctx, &categories, false)
		resp.Diagnostics.Append(diags...)
		
		tflog.Debug(ctx, "After parsing categories", map[string]any{
			"count": len(categories),
			"hasError": diags.HasError(),
			"categories": fmt.Sprintf("%v", categories),
		})
		
		if !diags.HasError() && len(categories) > 0 {
			// Filter out empty/invalid UUIDs
			validCategories := make([]string, 0, len(categories))
			for _, cat := range categories {
				if cat != "" {
					validCategories = append(validCategories, cat)
					tflog.Debug(ctx, "Valid category UUID", map[string]any{"uuid": cat})
				} else {
					tflog.Warn(ctx, "Skipping empty category UUID")
				}
			}
			
			if len(validCategories) > 0 {
				categoryStr := strings.Join(validCategories, ",")
				ruleData["rule"].(map[string]interface{})["category"] = categoryStr
				tflog.Info(ctx, "Setting categories on rule", map[string]any{
					"categories": categoryStr,
					"count": len(validCategories),
				})
			} else {
				tflog.Warn(ctx, "No valid category UUIDs found")
			}
		} else if diags.HasError() {
			tflog.Error(ctx, "Error parsing categories", map[string]any{
				"errors": fmt.Sprintf("%v", diags.Errors()),
			})
		} else {
			tflog.Warn(ctx, "Categories list is empty")
		}
	} else {
		tflog.Debug(ctx, "No categories field", map[string]any{
			"isNull": data.Categories.IsNull(),
			"isUnknown": data.Categories.IsUnknown(),
		})
	}

	// Make API call to create rule
	jsonData, err := json.Marshal(ruleData)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to marshal rule data: %s", err))
		return
	}

	url := fmt.Sprintf("%s/api/firewall/filter/addRule", r.client.Host)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, strings.NewReader(string(jsonData)))
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create request: %s", err))
		return
	}

	httpReq.SetBasicAuth(r.client.ApiKey, r.client.ApiSecret)
	httpReq.Header.Set("Content-Type", "application/json")

	tflog.Debug(ctx, "Creating firewall rule", map[string]any{"url": url})

	httpResp, err := r.client.client.Do(httpReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create rule: %s", err))
		return
	}
	defer httpResp.Body.Close()

	body, err := io.ReadAll(httpResp.Body)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read response: %s", err))
		return
	}

	if httpResp.StatusCode != 200 {
		resp.Diagnostics.AddError("API Error", fmt.Sprintf("API returned status %d: %s", httpResp.StatusCode, string(body)))
		return
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to parse response: %s\nRaw response: %s", err, string(body)))
		return
	}

	// Log the full response for debugging (debug level, not warning)
	if resultJSON, err := json.MarshalIndent(result, "", "  "); err == nil {
		tflog.Debug(ctx, "API Response", map[string]any{"response": string(resultJSON)})
	}

	// Try to extract UUID from various possible response formats
	var uuid string
	if uuidVal, ok := result["uuid"].(string); ok {
		uuid = uuidVal
	} else if uuidVal, ok := result["id"].(string); ok {
		uuid = uuidVal
	} else if resultVal, ok := result["result"].(string); ok {
		uuid = resultVal
	} else if ruleData, ok := result["rule"].(map[string]interface{}); ok {
		if uuidVal, ok := ruleData["uuid"].(string); ok {
			uuid = uuidVal
		}
	}

	if uuid != "" {
		data.ID = types.StringValue(uuid)
	} else {
		resp.Diagnostics.AddError("API Error", fmt.Sprintf("No UUID found in API response. Full response: %s", string(body)))
		return
	}

	// Apply the configuration
	applyURL := fmt.Sprintf("%s/api/firewall/filter/apply", r.client.Host)
	applyReq, _ := http.NewRequestWithContext(ctx, "POST", applyURL, nil)
	applyReq.SetBasicAuth(r.client.ApiKey, r.client.ApiSecret)
	r.client.client.Do(applyReq)

	tflog.Trace(ctx, "created firewall rule resource")

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *FirewallRuleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data FirewallRuleResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Get rule by UUID
	url := fmt.Sprintf("%s/api/firewall/filter/getRule/%s", r.client.Host, data.ID.ValueString())
	httpReq, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create request: %s", err))
		return
	}

	httpReq.SetBasicAuth(r.client.ApiKey, r.client.ApiSecret)

	httpResp, err := r.client.client.Do(httpReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read rule: %s", err))
		return
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode == 404 {
		resp.State.RemoveResource(ctx)
		return
	}

	if httpResp.StatusCode != 200 {
		resp.Diagnostics.AddError("API Error", fmt.Sprintf("API returned status %d", httpResp.StatusCode))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *FirewallRuleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data FirewallRuleResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Similar to Create, but use setRule endpoint with UUID
	ruleData := map[string]interface{}{
		"rule": map[string]interface{}{
			"description": data.Description.ValueString(),
			"source_net":  data.SourceNet.ValueString(),
			"destination_net": data.DestNet.ValueString(),
			"protocol":    data.Protocol.ValueString(),
		},
	}

	// Add sequence if provided
	if !data.Sequence.IsNull() {
		ruleData["rule"].(map[string]interface{})["sequence"] = fmt.Sprintf("%d", data.Sequence.ValueInt64())
	}

	// Add optional fields (same as Create)
	if !data.Interface.IsNull() {
		ruleData["rule"].(map[string]interface{})["interface"] = data.Interface.ValueString()
	}
	if !data.Direction.IsNull() {
		ruleData["rule"].(map[string]interface{})["direction"] = data.Direction.ValueString()
	}
	if !data.IPProtocol.IsNull() {
		ruleData["rule"].(map[string]interface{})["ipprotocol"] = data.IPProtocol.ValueString()
	}
	if !data.SourcePort.IsNull() {
		ruleData["rule"].(map[string]interface{})["source_port"] = data.SourcePort.ValueString()
	}
	if !data.DestPort.IsNull() {
		ruleData["rule"].(map[string]interface{})["destination_port"] = data.DestPort.ValueString()
	}
	if !data.Gateway.IsNull() {
		ruleData["rule"].(map[string]interface{})["gateway"] = data.Gateway.ValueString()
	}
	if !data.Action.IsNull() {
		ruleData["rule"].(map[string]interface{})["action"] = data.Action.ValueString()
	}
	if !data.Enabled.IsNull() {
		if data.Enabled.ValueBool() {
			ruleData["rule"].(map[string]interface{})["enabled"] = "1"
		} else {
			ruleData["rule"].(map[string]interface{})["enabled"] = "0"
		}
	}
	if !data.Log.IsNull() {
		if data.Log.ValueBool() {
			ruleData["rule"].(map[string]interface{})["log"] = "1"
		} else {
			ruleData["rule"].(map[string]interface{})["log"] = "0"
		}
	}
	if !data.Quick.IsNull() {
		if data.Quick.ValueBool() {
			ruleData["rule"].(map[string]interface{})["quick"] = "1"
		} else {
			ruleData["rule"].(map[string]interface{})["quick"] = "0"
		}
	}
	if !data.Invert.IsNull() {
		if data.Invert.ValueBool() {
			ruleData["rule"].(map[string]interface{})["destination_not"] = "1"
		} else {
			ruleData["rule"].(map[string]interface{})["destination_not"] = "0"
		}
	}
	// Handle deprecated but still functional destination_not field
	if !data.DestinationNot.IsNull() {
		if data.DestinationNot.ValueBool() {
			ruleData["rule"].(map[string]interface{})["destination_not"] = "1"
		} else {
			ruleData["rule"].(map[string]interface{})["destination_not"] = "0"
		}
	}
	// Handle deprecated but still functional source_not field
	if !data.SourceNot.IsNull() {
		if data.SourceNot.ValueBool() {
			ruleData["rule"].(map[string]interface{})["source_not"] = "1"
		} else{
			ruleData["rule"].(map[string]interface{})["source_not"] = "0"
		}
	}
	if !data.Categories.IsNull() {
		var categories []string
		resp.Diagnostics.Append(data.Categories.ElementsAs(ctx, &categories, false)...)
		if !resp.Diagnostics.HasError() && len(categories) > 0 {
			ruleData["rule"].(map[string]interface{})["category"] = strings.Join(categories, ",")
		}
	}

	jsonData, _ := json.Marshal(ruleData)

	url := fmt.Sprintf("%s/api/firewall/filter/setRule/%s", r.client.Host, data.ID.ValueString())
	httpReq, _ := http.NewRequestWithContext(ctx, "POST", url, strings.NewReader(string(jsonData)))
	httpReq.SetBasicAuth(r.client.ApiKey, r.client.ApiSecret)
	httpReq.Header.Set("Content-Type", "application/json")

	httpResp, err := r.client.client.Do(httpReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update rule: %s", err))
		return
	}
	defer httpResp.Body.Close()

	// Apply the configuration
	applyURL := fmt.Sprintf("%s/api/firewall/filter/apply", r.client.Host)
	applyReq, _ := http.NewRequestWithContext(ctx, "POST", applyURL, nil)
	applyReq.SetBasicAuth(r.client.ApiKey, r.client.ApiSecret)
	r.client.client.Do(applyReq)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *FirewallRuleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data FirewallRuleResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	url := fmt.Sprintf("%s/api/firewall/filter/delRule/%s", r.client.Host, data.ID.ValueString())
	httpReq, _ := http.NewRequestWithContext(ctx, "POST", url, nil)
	httpReq.SetBasicAuth(r.client.ApiKey, r.client.ApiSecret)

	httpResp, err := r.client.client.Do(httpReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete rule: %s", err))
		return
	}
	defer httpResp.Body.Close()

	// Apply the configuration
	applyURL := fmt.Sprintf("%s/api/firewall/filter/apply", r.client.Host)
	applyReq, _ := http.NewRequestWithContext(ctx, "POST", applyURL, nil)
	applyReq.SetBasicAuth(r.client.ApiKey, r.client.ApiSecret)
	r.client.client.Do(applyReq)
}

func (r *FirewallRuleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}