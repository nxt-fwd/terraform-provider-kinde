package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/nxt-fwd/kinde-go"
	"github.com/nxt-fwd/kinde-go/api/connections"
)

var _ datasource.DataSource = &ConnectionsDataSource{}

func NewConnectionsDataSource() datasource.DataSource {
	return &ConnectionsDataSource{}
}

type ConnectionsDataSource struct {
	client *connections.Client
}

type ConnectionsDataSourceModel struct {
	Filter      types.String      `tfsdk:"filter"`
	Connections []ConnectionModel `tfsdk:"connections"`
}

type ConnectionModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	DisplayName types.String `tfsdk:"display_name"`
	Strategy    types.String `tfsdk:"strategy"`
}

func (d *ConnectionsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_connections"
}

func (d *ConnectionsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Use this data source to list available connections.",

		Attributes: map[string]schema.Attribute{
			"filter": schema.StringAttribute{
				MarkdownDescription: "Filter connections by type. Valid values are: `builtin`, `custom`, `all`. Defaults to `all`.",
				Optional:            true,
			},
			"connections": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed: true,
						},
						"name": schema.StringAttribute{
							Computed: true,
						},
						"display_name": schema.StringAttribute{
							Computed: true,
						},
						"strategy": schema.StringAttribute{
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func (d *ConnectionsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*kinde.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *kinde.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	d.client = client.Connections
}

func isBuiltinStrategy(strategy string) bool {
	switch strategy {
	case string(connections.StrategyEmailPassword),
		string(connections.StrategyEmailOTP),
		string(connections.StrategyPhoneOTP),
		string(connections.StrategyUsernamePassword),
		string(connections.StrategyUsernameOTP):
		return true
	default:
		return false
	}
}

func (d *ConnectionsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ConnectionsDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get all connections
	conns, err := d.client.List(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read connections, got error: %s", err))
		return
	}

	// Filter connections based on the filter parameter
	filter := "all"
	if !data.Filter.IsNull() {
		filter = data.Filter.ValueString()
	}

	var filteredConns []connections.Connection
	switch filter {
	case "builtin":
		for _, conn := range conns {
			if isBuiltinStrategy(conn.Strategy) {
				filteredConns = append(filteredConns, conn)
			}
		}
	case "custom":
		for _, conn := range conns {
			if !isBuiltinStrategy(conn.Strategy) {
				filteredConns = append(filteredConns, conn)
			}
		}
	default:
		filteredConns = conns
	}

	// Convert to model
	data.Connections = make([]ConnectionModel, len(filteredConns))
	for i, conn := range filteredConns {
		data.Connections[i] = ConnectionModel{
			ID:          types.StringValue(conn.ID),
			Name:        types.StringValue(conn.Name),
			DisplayName: types.StringValue(conn.DisplayName),
			Strategy:    types.StringValue(conn.Strategy),
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
