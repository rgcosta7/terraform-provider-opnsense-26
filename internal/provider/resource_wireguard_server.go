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
)

var _ resource.Resource = &WireguardServerResource{}
var _ resource.ResourceWithImportState = &WireguardServerResource{}

func NewWireguardServerResource() resource.Resource {
	return &WireguardServerResource{}
}

type WireguardServerResource struct {
	client *Client
}

type WireguardServerResourceModel struct {
	ID            types.String `tfsdk:"id"`
	Name          types.String `tfsdk:"name"`
	Enabled       types.Bool   `tfsdk:"enabled"`
	PublicKey     types.String `tfsdk:"public_key"`
	PrivateKey    types.String `tfsdk:"private_key"`
	ListenPort    types.Int64  `tfsdk:"listen_port"`
	TunnelAddr    types.String `tfsdk:"tunnel_address"`
	Peers         types.List   `tfsdk:"peers"`
	DisableRoutes types.Bool   `tfsdk:"disable_routes"`
	DNS           types.String `tfsdk:"dns"`
	MTU           types.Int64  `tfsdk:"mtu"`
	Gateway       types.String `tfsdk:"gateway"`
}

func (r *WireguardServerResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_wireguard_server"
}

func (r *WireguardServerResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages WireGuard server instances in OPNsense",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Server UUID",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the WireGuard server instance",
				Required:            true,
			},
			"enabled": schema.BoolAttribute{
				MarkdownDescription: "Whether the server is enabled",
				Optional:            true,
			},
			"public_key": schema.StringAttribute{
				MarkdownDescription: "Server public key",
				Computed:            true,
			},
			"private_key": schema.StringAttribute{
				MarkdownDescription: "Server private key (auto-generated if not provided)",
				Optional:            true,
				Computed:            true,
				Sensitive:           true,
			},
			"listen_port": schema.Int64Attribute{
				MarkdownDescription: "UDP port to listen on",
				Required:            true,
			},
			"tunnel_address": schema.StringAttribute{
				MarkdownDescription: "Tunnel address in CIDR notation (e.g., 10.10.10.1/24)",
				Required:            true,
			},
			"peers": schema.ListAttribute{
				MarkdownDescription: "List of peer UUIDs",
				Optional:            true,
				ElementType:         types.StringType,
			},
			"disable_routes": schema.BoolAttribute{
				MarkdownDescription: "Disable automatic route creation",
				Optional:            true,
			},
			"dns": schema.StringAttribute{
				MarkdownDescription: "DNS servers for clients (comma-separated)",
				Optional:            true,
			},
			"mtu": schema.Int64Attribute{
				MarkdownDescription: "MTU for the tunnel interface",
				Optional:            true,
			},
			"gateway": schema.StringAttribute{
				MarkdownDescription: "Gateway IP address for the tunnel",
				Optional:            true,
			},
		},
	}
}

func (r *WireguardServerResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *WireguardServerResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data WireguardServerResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	serverData := map[string]interface{}{
		"server": map[string]interface{}{
			"name":           data.Name.ValueString(),
			"port":           fmt.Sprintf("%d", data.ListenPort.ValueInt64()),
			"tunneladdress":  data.TunnelAddr.ValueString(),
		},
	}

	if !data.Enabled.IsNull() {
		if data.Enabled.ValueBool() {
			serverData["server"].(map[string]interface{})["enabled"] = "1"
		} else {
			serverData["server"].(map[string]interface{})["enabled"] = "0"
		}
	} else {
		serverData["server"].(map[string]interface{})["enabled"] = "1"
	}

	if !data.PrivateKey.IsNull() {
		serverData["server"].(map[string]interface{})["privkey"] = data.PrivateKey.ValueString()
	}

	if !data.DisableRoutes.IsNull() && data.DisableRoutes.ValueBool() {
		serverData["server"].(map[string]interface{})["disableroutes"] = "1"
	}

	if !data.Peers.IsNull() {
		var peers []string
		resp.Diagnostics.Append(data.Peers.ElementsAs(ctx, &peers, false)...)
		serverData["server"].(map[string]interface{})["peers"] = strings.Join(peers, ",")
	}

	if !data.DNS.IsNull() {
		serverData["server"].(map[string]interface{})["dns"] = data.DNS.ValueString()
	}

	if !data.MTU.IsNull() {
		serverData["server"].(map[string]interface{})["mtu"] = fmt.Sprintf("%d", data.MTU.ValueInt64())
	}

	if !data.Gateway.IsNull() {
		serverData["server"].(map[string]interface{})["gateway"] = data.Gateway.ValueString()
	}

	jsonData, _ := json.Marshal(serverData)

	url := fmt.Sprintf("%s/api/wireguard/server/add_server", r.client.Host)
	httpReq, _ := http.NewRequestWithContext(ctx, "POST", url, strings.NewReader(string(jsonData)))
	httpReq.SetBasicAuth(r.client.ApiKey, r.client.ApiSecret)
	httpReq.Header.Set("Content-Type", "application/json")

	httpResp, err := r.client.client.Do(httpReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create server: %s", err))
		return
	}
	defer httpResp.Body.Close()

	body, _ := io.ReadAll(httpResp.Body)

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to parse response: %s", err))
		return
	}

	if uuid, ok := result["uuid"].(string); ok {
		data.ID = types.StringValue(uuid)
	} else {
		resp.Diagnostics.AddError("API Error", "No UUID returned from API")
		return
	}

	// Read back to get generated keys
	r.readServerKeys(ctx, &data)

	// Apply configuration
	applyURL := fmt.Sprintf("%s/api/wireguard/service/reconfigure", r.client.Host)
	applyReq, _ := http.NewRequestWithContext(ctx, "POST", applyURL, nil)
	applyReq.SetBasicAuth(r.client.ApiKey, r.client.ApiSecret)
	r.client.client.Do(applyReq)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *WireguardServerResource) readServerKeys(ctx context.Context, data *WireguardServerResourceModel) {
	url := fmt.Sprintf("%s/api/wireguard/server/get_server/%s", r.client.Host, data.ID.ValueString())
	httpReq, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
	httpReq.SetBasicAuth(r.client.ApiKey, r.client.ApiSecret)

	httpResp, err := r.client.client.Do(httpReq)
	if err != nil {
		return
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != 200 {
		return
	}

	body, _ := io.ReadAll(httpResp.Body)
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return
	}

	if server, ok := result["server"].(map[string]interface{}); ok {
		if pubkey, ok := server["pubkey"].(string); ok {
			data.PublicKey = types.StringValue(pubkey)
		}
		if privkey, ok := server["privkey"].(string); ok && data.PrivateKey.IsNull() {
			data.PrivateKey = types.StringValue(privkey)
		}
	}
}

func (r *WireguardServerResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data WireguardServerResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	url := fmt.Sprintf("%s/api/wireguard/server/get_server/%s", r.client.Host, data.ID.ValueString())
	httpReq, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
	httpReq.SetBasicAuth(r.client.ApiKey, r.client.ApiSecret)

	httpResp, err := r.client.client.Do(httpReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read server: %s", err))
		return
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode == 404 {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *WireguardServerResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data WireguardServerResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	serverData := map[string]interface{}{
		"server": map[string]interface{}{
			"name":           data.Name.ValueString(),
			"port":           fmt.Sprintf("%d", data.ListenPort.ValueInt64()),
			"tunneladdress":  data.TunnelAddr.ValueString(),
		},
	}

	if !data.Enabled.IsNull() {
		if data.Enabled.ValueBool() {
			serverData["server"].(map[string]interface{})["enabled"] = "1"
		} else {
			serverData["server"].(map[string]interface{})["enabled"] = "0"
		}
	}

	if !data.PrivateKey.IsNull() {
		serverData["server"].(map[string]interface{})["privkey"] = data.PrivateKey.ValueString()
	}

	if !data.DisableRoutes.IsNull() && data.DisableRoutes.ValueBool() {
		serverData["server"].(map[string]interface{})["disableroutes"] = "1"
	}

	if !data.Peers.IsNull() {
		var peers []string
		resp.Diagnostics.Append(data.Peers.ElementsAs(ctx, &peers, false)...)
		serverData["server"].(map[string]interface{})["peers"] = strings.Join(peers, ",")
	}

	if !data.DNS.IsNull() {
		serverData["server"].(map[string]interface{})["dns"] = data.DNS.ValueString()
	}

	if !data.MTU.IsNull() {
		serverData["server"].(map[string]interface{})["mtu"] = fmt.Sprintf("%d", data.MTU.ValueInt64())
	}

	if !data.Gateway.IsNull() {
		serverData["server"].(map[string]interface{})["gateway"] = data.Gateway.ValueString()
	}

	jsonData, _ := json.Marshal(serverData)

	url := fmt.Sprintf("%s/api/wireguard/server/set_server/%s", r.client.Host, data.ID.ValueString())
	httpReq, _ := http.NewRequestWithContext(ctx, "POST", url, strings.NewReader(string(jsonData)))
	httpReq.SetBasicAuth(r.client.ApiKey, r.client.ApiSecret)
	httpReq.Header.Set("Content-Type", "application/json")

	httpResp, err := r.client.client.Do(httpReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update server: %s", err))
		return
	}
	defer httpResp.Body.Close()

	// Apply configuration
	applyURL := fmt.Sprintf("%s/api/wireguard/service/reconfigure", r.client.Host)
	applyReq, _ := http.NewRequestWithContext(ctx, "POST", applyURL, nil)
	applyReq.SetBasicAuth(r.client.ApiKey, r.client.ApiSecret)
	r.client.client.Do(applyReq)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *WireguardServerResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data WireguardServerResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	url := fmt.Sprintf("%s/api/wireguard/server/del_server/%s", r.client.Host, data.ID.ValueString())
	httpReq, _ := http.NewRequestWithContext(ctx, "POST", url, nil)
	httpReq.SetBasicAuth(r.client.ApiKey, r.client.ApiSecret)

	httpResp, err := r.client.client.Do(httpReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete server: %s", err))
		return
	}
	defer httpResp.Body.Close()

	// Apply configuration
	applyURL := fmt.Sprintf("%s/api/wireguard/service/reconfigure", r.client.Host)
	applyReq, _ := http.NewRequestWithContext(ctx, "POST", applyURL, nil)
	applyReq.SetBasicAuth(r.client.ApiKey, r.client.ApiSecret)
	r.client.client.Do(applyReq)
}

func (r *WireguardServerResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
