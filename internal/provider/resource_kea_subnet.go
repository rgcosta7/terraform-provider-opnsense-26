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

var _ resource.Resource = &KeaSubnetResource{}
var _ resource.ResourceWithImportState = &KeaSubnetResource{}

func NewKeaSubnetResource() resource.Resource {
	return &KeaSubnetResource{}
}

type KeaSubnetResource struct {
	client *Client
}

type KeaSubnetResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Subnet      types.String `tfsdk:"subnet"`
	Pools       types.String `tfsdk:"pools"`
	Option      types.String `tfsdk:"option_data"`
	Description types.String `tfsdk:"description"`
}

func (r *KeaSubnetResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_kea_subnet"
}

func (r *KeaSubnetResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages Kea DHCP subnets in OPNsense",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Subnet UUID",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"subnet": schema.StringAttribute{
				MarkdownDescription: "Subnet in CIDR notation (e.g., 192.168.1.0/24)",
				Required:            true,
			},
			"pools": schema.StringAttribute{
				MarkdownDescription: "IP address pools (comma-separated ranges, e.g., '192.168.1.100-192.168.1.200')",
				Optional:            true,
			},
			"option_data": schema.StringAttribute{
				MarkdownDescription: "DHCP options data",
				Optional:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Description of the subnet",
				Optional:            true,
			},
		},
	}
}

func (r *KeaSubnetResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *KeaSubnetResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data KeaSubnetResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	subnetData := map[string]interface{}{
		"subnet4": map[string]interface{}{
			"subnet4": data.Subnet.ValueString(),
		},
	}

	if !data.Pools.IsNull() {
		subnetData["subnet4"].(map[string]interface{})["pools"] = data.Pools.ValueString()
	}
	if !data.Option.IsNull() {
		subnetData["subnet4"].(map[string]interface{})["option_data"] = data.Option.ValueString()
	}
	if !data.Description.IsNull() {
		subnetData["subnet4"].(map[string]interface{})["description"] = data.Description.ValueString()
	}

	jsonData, _ := json.Marshal(subnetData)

	// Log the request for debugging
	tflog.Debug(ctx, "Creating Kea subnet", map[string]any{
		"endpoint": "/api/kea/dhcpv4/add_subnet",
		"request":  string(jsonData),
	})

	url := fmt.Sprintf("%s/api/kea/dhcpv4/add_subnet", r.client.Host)
	httpReq, _ := http.NewRequestWithContext(ctx, "POST", url, strings.NewReader(string(jsonData)))
	httpReq.SetBasicAuth(r.client.ApiKey, r.client.ApiSecret)
	httpReq.Header.Set("Content-Type", "application/json")

	httpResp, err := r.client.client.Do(httpReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create subnet: %s", err))
		return
	}
	defer httpResp.Body.Close()

	body, err := io.ReadAll(httpResp.Body)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read response: %s", err))
		return
	}

	// Log raw response
	tflog.Debug(ctx, "Kea API Response", map[string]any{
		"status_code": httpResp.StatusCode,
		"body":        string(body),
	})

	if httpResp.StatusCode != 200 {
		resp.Diagnostics.AddError("API Error", fmt.Sprintf("API returned status %d: %s", httpResp.StatusCode, string(body)))
		return
	}

	// Check if response is empty or just whitespace
	if len(strings.TrimSpace(string(body))) == 0 {
		resp.Diagnostics.AddError("API Error", "API returned empty response")
		return
	}

	// Try to determine what kind of response we got
	firstChar := strings.TrimSpace(string(body))[0]
	
	if firstChar == '[' {
		// It's an array response - likely validation errors or empty response
		var resultArray []interface{}
		if err := json.Unmarshal(body, &resultArray); err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to parse array response: %s\nRaw response: %s", err, string(body)))
			return
		}
		
		if len(resultArray) == 0 {
			resp.Diagnostics.AddError(
				"Kea DHCP API Error",
				"API returned empty array [].\n\n"+
				"This typically means:\n"+
				"1. The Kea DHCP plugin is not installed or enabled in OPNsense\n"+
				"2. The API endpoint doesn't exist (wrong OPNsense version)\n"+
				"3. Request validation failed\n\n"+
				"To fix:\n"+
				"- In OPNsense GUI: System > Firmware > Plugins\n"+
				"- Install 'os-kea-dhcp' plugin if not already installed\n"+
				"- Check Services > Kea DHCPv4 to ensure it's configured\n\n"+
				"Request sent: "+string(jsonData),
			)
			return
		}
		
		// Try to extract error messages from array
		var errorMessages []string
		for _, item := range resultArray {
			if errMap, ok := item.(map[string]interface{}); ok {
				if msg, ok := errMap["message"].(string); ok {
					errorMessages = append(errorMessages, msg)
				}
			}
		}
		
		if len(errorMessages) > 0 {
			resp.Diagnostics.AddError("API Validation Error", fmt.Sprintf("API returned errors: %s", strings.Join(errorMessages, ", ")))
		} else {
			resp.Diagnostics.AddError("API Error", fmt.Sprintf("API returned unexpected array response: %s", string(body)))
		}
		return
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to parse response: %s\nRaw response: %s", err, string(body)))
		return
	}

	// Check for validation errors in result
	if resultStatus, ok := result["result"].(string); ok && resultStatus == "failed" {
		var errorMsgs []string
		if validations, ok := result["validations"].(map[string]interface{}); ok {
			for field, errors := range validations {
				if errList, ok := errors.([]interface{}); ok {
					for _, err := range errList {
						if errStr, ok := err.(string); ok {
							errorMsgs = append(errorMsgs, fmt.Sprintf("%s: %s", field, errStr))
						}
					}
				} else if errStr, ok := errors.(string); ok {
					errorMsgs = append(errorMsgs, fmt.Sprintf("%s: %s", field, errStr))
				}
			}
		}
		if len(errorMsgs) > 0 {
			resp.Diagnostics.AddError("API Validation Failed", fmt.Sprintf("Validation errors:\n- %s", strings.Join(errorMsgs, "\n- ")))
		} else {
			resp.Diagnostics.AddError("API Error", fmt.Sprintf("API returned failed status: %s", string(body)))
		}
		return
	}

	// Log the full response for debugging
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
	} else if subnetData, ok := result["subnet4"].(map[string]interface{}); ok {
		if uuidVal, ok := subnetData["uuid"].(string); ok {
			uuid = uuidVal
		}
	}

	if uuid != "" {
		data.ID = types.StringValue(uuid)
	} else {
		resp.Diagnostics.AddError("API Error", fmt.Sprintf("No UUID found in API response. Response: %s", string(body)))
		return
	}

	// Apply configuration
	applyURL := fmt.Sprintf("%s/api/kea/service/reconfigure", r.client.Host)
	applyReq, _ := http.NewRequestWithContext(ctx, "POST", applyURL, nil)
	applyReq.SetBasicAuth(r.client.ApiKey, r.client.ApiSecret)
	r.client.client.Do(applyReq)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *KeaSubnetResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data KeaSubnetResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	url := fmt.Sprintf("%s/api/kea/dhcpv4/get_subnet/%s", r.client.Host, data.ID.ValueString())
	httpReq, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
	httpReq.SetBasicAuth(r.client.ApiKey, r.client.ApiSecret)

	httpResp, err := r.client.client.Do(httpReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read subnet: %s", err))
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

	body, err := io.ReadAll(httpResp.Body)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read response: %s", err))
		return
	}

	// Log raw response
	tflog.Debug(ctx, "Kea subnet Read response", map[string]any{
		"status_code": httpResp.StatusCode,
		"body":        string(body),
	})

	// Check if response is empty
	if len(strings.TrimSpace(string(body))) == 0 {
		// Empty response might mean the subnet doesn't exist anymore
		resp.State.RemoveResource(ctx)
		return
	}

	// Try to determine what kind of response we got
	firstChar := strings.TrimSpace(string(body))[0]
	
	if firstChar == '[' {
		// Array response - likely means resource doesn't exist or error
		tflog.Warn(ctx, "Kea subnet returned array, removing from state", map[string]any{
			"id": data.ID.ValueString(),
		})
		resp.State.RemoveResource(ctx)
		return
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to parse response: %s\nRaw response: %s", err, string(body)))
		return
	}

	// Parse the subnet data from the response
	if subnetData, ok := result["subnet4"].(map[string]interface{}); ok {
		if subnet, ok := subnetData["subnet"].(string); ok {
			data.Subnet = types.StringValue(subnet)
		}
		if pools, ok := subnetData["pools"].(string); ok {
			data.Pools = types.StringValue(pools)
		}
		if optionData, ok := subnetData["option_data"].(string); ok {
			data.Option = types.StringValue(optionData)
		}
		if description, ok := subnetData["description"].(string); ok {
			data.Description = types.StringValue(description)
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *KeaSubnetResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data KeaSubnetResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	subnetData := map[string]interface{}{
		"subnet4": map[string]interface{}{
			"subnet4": data.Subnet.ValueString(),
		},
	}

	if !data.Pools.IsNull() {
		subnetData["subnet4"].(map[string]interface{})["pools"] = data.Pools.ValueString()
	}
	if !data.Option.IsNull() {
		subnetData["subnet4"].(map[string]interface{})["option_data"] = data.Option.ValueString()
	}
	if !data.Description.IsNull() {
		subnetData["subnet4"].(map[string]interface{})["description"] = data.Description.ValueString()
	}

	jsonData, _ := json.Marshal(subnetData)

	url := fmt.Sprintf("%s/api/kea/dhcpv4/set_subnet/%s", r.client.Host, data.ID.ValueString())
	httpReq, _ := http.NewRequestWithContext(ctx, "POST", url, strings.NewReader(string(jsonData)))
	httpReq.SetBasicAuth(r.client.ApiKey, r.client.ApiSecret)
	httpReq.Header.Set("Content-Type", "application/json")

	httpResp, err := r.client.client.Do(httpReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update subnet: %s", err))
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

	// Apply configuration
	applyURL := fmt.Sprintf("%s/api/kea/service/reconfigure", r.client.Host)
	applyReq, _ := http.NewRequestWithContext(ctx, "POST", applyURL, nil)
	applyReq.SetBasicAuth(r.client.ApiKey, r.client.ApiSecret)
	r.client.client.Do(applyReq)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *KeaSubnetResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data KeaSubnetResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	url := fmt.Sprintf("%s/api/kea/dhcpv4/del_subnet/%s", r.client.Host, data.ID.ValueString())
	httpReq, _ := http.NewRequestWithContext(ctx, "POST", url, nil)
	httpReq.SetBasicAuth(r.client.ApiKey, r.client.ApiSecret)

	httpResp, err := r.client.client.Do(httpReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete subnet: %s", err))
		return
	}
	defer httpResp.Body.Close()

	// Apply configuration
	applyURL := fmt.Sprintf("%s/api/kea/service/reconfigure", r.client.Host)
	applyReq, _ := http.NewRequestWithContext(ctx, "POST", applyURL, nil)
	applyReq.SetBasicAuth(r.client.ApiKey, r.client.ApiSecret)
	r.client.client.Do(applyReq)
}

func (r *KeaSubnetResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}