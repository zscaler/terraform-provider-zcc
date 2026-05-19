package datasources

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/zscaler/zscaler-sdk-go/v3/zscaler/zcc/services/process_based_apps"

	"github.com/zscaler/terraform-provider-zcc/internal/client"
)

var (
	_ datasource.DataSource              = &ProcessBasedAppsDataSource{}
	_ datasource.DataSourceWithConfigure = &ProcessBasedAppsDataSource{}
)

func NewProcessBasedAppsDataSource() datasource.DataSource {
	return &ProcessBasedAppsDataSource{}
}

type ProcessBasedAppsDataSource struct {
	client *client.Client
}

type ProcessBasedAppsDataSourceModel struct {
	ID                 types.String `tfsdk:"id"`
	Name               types.String `tfsdk:"name"`
	AppName            types.String `tfsdk:"app_name"`
	FileNames          types.List   `tfsdk:"file_names"`
	FilePaths          types.List   `tfsdk:"file_paths"`
	MatchingCriteria   types.Int64  `tfsdk:"matching_criteria"`
	SignaturePayload   types.String `tfsdk:"signature_payload"`
	CertificatePayload types.String `tfsdk:"certificate_payload"`
	CreatedBy          types.String `tfsdk:"created_by"`
	EditedBy           types.String `tfsdk:"edited_by"`
	EditedTimestamp    types.String `tfsdk:"edited_timestamp"`
}

func (d *ProcessBasedAppsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_process_based_apps"
}

func (d *ProcessBasedAppsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Retrieves a ZCC process-based app by ID or name.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the process-based app.",
				Optional:    true,
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "The name of the process-based app (used for lookup).",
				Optional:    true,
			},
			"app_name": schema.StringAttribute{Computed: true},
			"file_names": schema.ListAttribute{
				ElementType: types.StringType,
				Computed:    true,
			},
			"file_paths": schema.ListAttribute{
				ElementType: types.StringType,
				Computed:    true,
			},
			"matching_criteria":   schema.Int64Attribute{Computed: true},
			"signature_payload":   schema.StringAttribute{Computed: true},
			"certificate_payload": schema.StringAttribute{Computed: true},
			"created_by":          schema.StringAttribute{Computed: true},
			"edited_by":           schema.StringAttribute{Computed: true},
			"edited_timestamp":    schema.StringAttribute{Computed: true},
		},
	}
}

func (d *ProcessBasedAppsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T", req.ProviderData),
		)
		return
	}

	d.client = c
}

func flattenStringSlice(items []string) types.List {
	if len(items) == 0 {
		return types.ListNull(types.StringType)
	}
	elements := make([]attr.Value, 0, len(items))
	for _, s := range items {
		elements = append(elements, types.StringValue(s))
	}
	list, _ := types.ListValue(types.StringType, elements)
	return list
}

func (d *ProcessBasedAppsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ProcessBasedAppsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if (data.ID.IsNull() || data.ID.ValueString() == "") && (data.Name.IsNull() || data.Name.ValueString() == "") {
		resp.Diagnostics.AddError("Missing Identifier", "Either id or name must be specified")
		return
	}

	service := d.client.Service

	var app *process_based_apps.ProcessBasedApp

	if !data.ID.IsNull() && data.ID.ValueString() != "" {
		id := data.ID.ValueString()
		tflog.Info(ctx, "Fetching process-based app by ID", map[string]any{"id": id})
		result, _, err := process_based_apps.GetByAppID(ctx, service, id)
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read process-based app: %v", err))
			return
		}
		app = result
	} else {
		name := data.Name.ValueString()
		tflog.Info(ctx, "Fetching process-based app by name", map[string]any{"name": name})
		result, _, err := process_based_apps.GetByName(ctx, service, name)
		if err != nil {
			resp.Diagnostics.AddError("Not Found", fmt.Sprintf("Process-based app with name '%s' not found: %v", name, err))
			return
		}
		app = result
	}

	model := ProcessBasedAppsDataSourceModel{
		ID:                 types.StringValue(strconv.Itoa(app.ID)),
		Name:               types.StringValue(app.AppName),
		AppName:            types.StringValue(app.AppName),
		FileNames:          flattenStringSlice(app.FileNames),
		FilePaths:          flattenStringSlice(app.FilePaths),
		MatchingCriteria:   types.Int64Value(int64(app.MatchingCriteria)),
		SignaturePayload:   types.StringValue(app.SignaturePayload),
		CertificatePayload: types.StringValue(app.CertificatePayload),
		CreatedBy:          types.StringValue(app.CreatedBy),
		EditedBy:           types.StringValue(app.EditedBy),
		EditedTimestamp:    types.StringValue(app.EditedTimestamp),
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}
