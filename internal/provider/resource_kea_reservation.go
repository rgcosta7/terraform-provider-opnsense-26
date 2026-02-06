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

var _ resource.Resource = &KeaReservationResource{}
var _ resource.ResourceWithImportState = &KeaReservationResource{}

func NewKeaReservationResource() resource.Resource {
	return &KeaReservationResource{}
}

type KeaReservationResource struct {
	client *Client
}

type KeaReservationResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Subnet      types.String `tfsdk:"subnet"`
	IPAddress   types.String `tfsdk:"ip_address"`
	HWAddress   types.String `tfsdk:"hw_address"`
	Hostname    types.String `tfsdk:"hostname"`
	Description types.String `tfsdk:"description"`
}

func (r *KeaReservationResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_kea_reservation"
}

func (r *KeaReservationResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages Kea DHCP reservations in OPNsense",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Reservation UUID",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"subnet": schema.StringAttribute{
				MarkdownDescription: "Subnet UUID this reservation belongs to",
				Required:            true,
			},
			"ip_address": schema.StringAttribute{
				MarkdownDescription: "Reserved IP address",
				Required:            true,
			},
			"hw_address": schema.StringAttribute{
				MarkdownDescription: "Hardware (MAC) address",
				Required:            true,
			},
			"hostname": schema.StringAttribute{
				MarkdownDescription: "Hostname for this reservation",
				Optional:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Description of the reservation",
				Optional:            true,
			},
		},
	}
}

func (r *KeaReservationResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *KeaReservationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data KeaReservationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	reservationData := map[string]interface{}{
		"reservation4": map[string]interface{}{
			"subnet":     data.Subnet.ValueString(),
			"ip_address": data.IPAddress.ValueString(),
			"hw_address": data.HWAddress.ValueString(),
		},
	}

	if !data.Hostname.IsNull() {
		reservationData["reservation4"].(map[string]interface{})["hostname"] = data.Hostname.ValueString()
	}
	if !data.Description.IsNull() {
		reservationData["reservation4"].(map[string]interface{})["description"] = data.Description.ValueString()
	}

	jsonData, _ := json.Marshal(reservationData)

	// Log the request for debugging
	tflog.Debug(ctx, "Creating Kea reservation", map[string]any{
		"endpoint": "/api/kea/dhcpv4/add_reservation",
		"request":  string(jsonData),
	})

	url := fmt.Sprintf("%s/api/kea/dhcpv4/add_reservation", r.client.Host)
	httpReq, _ := http.NewRequestWithContext(ctx, "POST", url, strings.NewReader(string(jsonData)))
	httpReq.SetBasicAuth(r.client.ApiKey, r.client.ApiSecret)
	httpReq.Header.Set("Content-Type", "application/json")

	httpResp, err := r.client.client.Do(httpReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create reservation: %s", err))
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
				"2. The subnet UUID referenced doesn't exist\n"+
				"3. Request validation failed\n\n"+
				"To fix:\n"+
				"- In OPNsense GUI: System > Firmware > Plugins\n"+
				"- Install 'os-kea-dhcp' plugin if not already installed\n"+
				"- Ensure the referenced subnet exists first\n"+
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
	} else if reservationData, ok := result["reservation4"].(map[string]interface{}); ok {
		if uuidVal, ok := reservationData["uuid"].(string); ok {
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

func (r *KeaReservationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data KeaReservationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	url := fmt.Sprintf("%s/api/kea/dhcpv4/get_reservation/%s", r.client.Host, data.ID.ValueString())
	httpReq, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
	httpReq.SetBasicAuth(r.client.ApiKey, r.client.ApiSecret)

	httpResp, err := r.client.client.Do(httpReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read reservation: %s", err))
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
	tflog.Debug(ctx, "Kea reservation Read response", map[string]any{
		"status_code": httpResp.StatusCode,
		"body":        string(body),
	})

	// Check if response is empty
	if len(strings.TrimSpace(string(body))) == 0 {
		// Empty response might mean the reservation doesn't exist anymore
		resp.State.RemoveResource(ctx)
		return
	}

	// Try to determine what kind of response we got
	firstChar := strings.TrimSpace(string(body))[0]
	
	if firstChar == '[' {
		// Array response - likely means resource doesn't exist or error
		tflog.Warn(ctx, "Kea reservation returned array, removing from state", map[string]any{
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

	// Parse the reservation data from the response
	if reservationData, ok := result["reservation4"].(map[string]interface{}); ok {
		if subnet, ok := reservationData["subnet"].(string); ok {
			data.Subnet = types.StringValue(subnet)
		}
		if ipAddress, ok := reservationData["ip_address"].(string); ok {
			data.IPAddress = types.StringValue(ipAddress)
		}
		if hwAddress, ok := reservationData["hw_address"].(string); ok {
			data.HWAddress = types.StringValue(hwAddress)
		}
		if hostname, ok := reservationData["hostname"].(string); ok {
			data.Hostname = types.StringValue(hostname)
		}
		if description, ok := reservationData["description"].(string); ok {
			data.Description = types.StringValue(description)
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *KeaReservationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data KeaReservationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	reservationData := map[string]interface{}{
		"reservation4": map[string]interface{}{
			"subnet":     data.Subnet.ValueString(),
			"ip_address": data.IPAddress.ValueString(),
			"hw_address": data.HWAddress.ValueString(),
		},
	}

	if !data.Hostname.IsNull() {
		reservationData["reservation4"].(map[string]interface{})["hostname"] = data.Hostname.ValueString()
	}
	if !data.Description.IsNull() {
		reservationData["reservation4"].(map[string]interface{})["description"] = data.Description.ValueString()
	}

	jsonData, _ := json.Marshal(reservationData)

	url := fmt.Sprintf("%s/api/kea/dhcpv4/set_reservation/%s", r.client.Host, data.ID.ValueString())
	httpReq, _ := http.NewRequestWithContext(ctx, "POST", url, strings.NewReader(string(jsonData)))
	httpReq.SetBasicAuth(r.client.ApiKey, r.client.ApiSecret)
	httpReq.Header.Set("Content-Type", "application/json")

	httpResp, err := r.client.client.Do(httpReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update reservation: %s", err))
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

func (r *KeaReservationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data KeaReservationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	url := fmt.Sprintf("%s/api/kea/dhcpv4/del_reservation/%s", r.client.Host, data.ID.ValueString())
	httpReq, _ := http.NewRequestWithContext(ctx, "POST", url, nil)
	httpReq.SetBasicAuth(r.client.ApiKey, r.client.ApiSecret)

	httpResp, err := r.client.client.Do(httpReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete reservation: %s", err))
		return
	}
	defer httpResp.Body.Close()

	// Apply configuration
	applyURL := fmt.Sprintf("%s/api/kea/service/reconfigure", r.client.Host)
	applyReq, _ := http.NewRequestWithContext(ctx, "POST", applyURL, nil)
	applyReq.SetBasicAuth(r.client.ApiKey, r.client.ApiSecret)
	r.client.client.Do(applyReq)
}

func (r *KeaReservationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}