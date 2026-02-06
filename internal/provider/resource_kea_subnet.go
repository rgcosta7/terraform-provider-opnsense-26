package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
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
	Option      types.Map    `tfsdk:"option_data"`
	AutoCollect types.Bool   `tfsdk:"auto_collect"`
	Description types.String `tfsdk:"description"`
}

func (r *KeaSubnetResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_kea_subnet"
}

func (r *KeaSubnetResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"subnet":      schema.StringAttribute{Required: true},
			"pools":       schema.StringAttribute{Optional: true},
			"description": schema.StringAttribute{Optional: true},
			"auto_collect": schema.BoolAttribute{
				Optional: true,
				Computed: true,
			},
			"option_data": schema.MapAttribute{
				ElementType: types.StringType,
				Optional:    true,
			},
		},
	}
}

func (r *KeaSubnetResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil { return }
	client, ok := req.ProviderData.(*Client)
	if !ok {
		resp.Diagnostics.AddError("Type Error", "Expected *Client")
		return
	}
	r.client = client
}

func (r *KeaSubnetResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data KeaSubnetResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() { return }

	payload := r.mapToPayload(ctx, &data)
	jsonData, _ := json.Marshal(payload)

	url := fmt.Sprintf("%s/api/kea/dhcpv4/add_subnet", r.client.Host)
	body := r.doRequest(ctx, "POST", url, jsonData, &resp.Diagnostics)
	if resp.Diagnostics.HasError() { return }

	var result map[string]interface{}
	json.Unmarshal(body, &result)

	// OPNsense returns UUID in 'uuid' field
	if uuid, ok := result["uuid"].(string); ok {
		data.ID = types.StringValue(uuid)
	}

	r.reconfigureService(ctx)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *KeaSubnetResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data KeaSubnetResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() { return }

	url := fmt.Sprintf("%s/api/kea/dhcpv4/get_subnet/%s", r.client.Host, data.ID.ValueString())
	body := r.doRequest(ctx, "GET", url, nil, &resp.Diagnostics)
	if resp.Diagnostics.HasError() || body == nil { return }

	var result map[string]interface{}
	json.Unmarshal(body, &result)

	if subnetData, ok := result["subnet4"].(map[string]interface{}); ok {
		data.Subnet = types.StringValue(subnetData["subnet"].(string))
		data.Pools = types.StringValue(subnetData["pools"].(string))
		data.Description = types.StringValue(subnetData["description"].(string))
		data.AutoCollect = types.BoolValue(subnetData["option_data_autocollect"].(string) == "1")
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *KeaSubnetResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data KeaSubnetResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() { return }

	payload := r.mapToPayload(ctx, &data)
	jsonData, _ := json.Marshal(payload)

	url := fmt.Sprintf("%s/api/kea/dhcpv4/set_subnet/%s", r.client.Host, data.ID.ValueString())
	r.doRequest(ctx, "POST", url, jsonData, &resp.Diagnostics)
	
	r.reconfigureService(ctx)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *KeaSubnetResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data KeaSubnetResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	url := fmt.Sprintf("%s/api/kea/dhcpv4/del_subnet/%s", r.client.Host, data.ID.ValueString())
	r.doRequest(ctx, "POST", url, nil, &resp.Diagnostics)
	r.reconfigureService(ctx)
}

func (r *KeaSubnetResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *KeaSubnetResource) mapToPayload(ctx context.Context, data *KeaSubnetResourceModel) map[string]interface{} {
	subnet4 := map[string]interface{}{
		"subnet": data.Subnet.ValueString(),
	}

	if !data.Pools.IsNull() { subnet4["pools"] = data.Pools.ValueString() }
	if !data.Description.IsNull() { subnet4["description"] = data.Description.ValueString() }
	
	// Convert bool to OPNsense string "0" or "1"
	if !data.AutoCollect.IsNull() && !data.AutoCollect.ValueBool() {
		subnet4["option_data_autocollect"] = "0"
	} else {
		subnet4["option_data_autocollect"] = "1"
	}

	// Handle the options map
	if !data.Option.IsNull() && !data.Option.IsUnknown() {
		var optionMap map[string]string
		data.Option.ElementsAs(ctx, &optionMap, false)
		
		formattedOptions := make(map[string]interface{})
		for k, v := range optionMap {
			key := strings.ReplaceAll(k, "-", "_")
			formattedOptions[key] = map[string]interface{}{
				"": map[string]interface{}{
					"value":    v,
					"selected": 1,
				},
			}
		}
		subnet4["option_data"] = formattedOptions
	}

	return map[string]interface{}{"subnet4": subnet4}
}

func (r *KeaSubnetResource) doRequest(ctx context.Context, method, url string, body []byte, diags *diag.Diagnostics) []byte {
	var reader io.Reader
	if body != nil { reader = strings.NewReader(string(body)) }
	req, _ := http.NewRequestWithContext(ctx, method, url, reader)
	req.SetBasicAuth(r.client.ApiKey, r.client.ApiSecret)
	req.Header.Set("Content-Type", "application/json")
	
	httpResp, err := r.client.client.Do(req)
	if err != nil {
		diags.AddError("Client Error", err.Error())
		return nil
	}
	defer httpResp.Body.Close()
	respBody, _ := io.ReadAll(httpResp.Body)
	
	if httpResp.StatusCode != 200 {
		diags.AddError("API Error", fmt.Sprintf("Status %d: %s", httpResp.StatusCode, string(respBody)))
		return nil
	}
	return respBody
}

func (r *KeaSubnetResource) reconfigureService(ctx context.Context) {
	url := fmt.Sprintf("%s/api/kea/service/reconfigure", r.client.Host)
	req, _ := http.NewRequestWithContext(ctx, "POST", url, nil)
	req.SetBasicAuth(r.client.ApiKey, r.client.ApiSecret)
	r.client.client.Do(req)
}