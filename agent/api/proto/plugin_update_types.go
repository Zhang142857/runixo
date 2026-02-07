// Hand-written types for PluginService and UpdateService
// These correspond to the proto definitions in proto/agent.proto
// but were not included in the protoc-generated code.
package proto

// ==================== Plugin System Types ====================

// PluginState enum
type PluginState int32

const (
	PluginState_PLUGIN_INSTALLED PluginState = 0
	PluginState_PLUGIN_ENABLED   PluginState = 1
	PluginState_PLUGIN_DISABLED  PluginState = 2
	PluginState_PLUGIN_ERROR     PluginState = 3
	PluginState_PLUGIN_UPDATING  PluginState = 4
)

// PluginType enum
type PluginType int32

const (
	PluginType_PLUGIN_CLIENT PluginType = 0
	PluginType_PLUGIN_AGENT  PluginType = 1
	PluginType_PLUGIN_HYBRID PluginType = 2
)

// PluginRequest message
type PluginRequest struct {
	PluginId string `protobuf:"bytes,1,opt,name=plugin_id,json=pluginId,proto3" json:"plugin_id,omitempty"`
}

func (x *PluginRequest) GetPluginId() string {
	if x != nil {
		return x.PluginId
	}
	return ""
}

// InstallPluginRequest message
type InstallPluginRequest struct {
	PluginId string `protobuf:"bytes,1,opt,name=plugin_id,json=pluginId,proto3" json:"plugin_id,omitempty"`
	Source   string `protobuf:"bytes,2,opt,name=source,proto3" json:"source,omitempty"`
	Url      string `protobuf:"bytes,3,opt,name=url,proto3" json:"url,omitempty"`
	Data     []byte `protobuf:"bytes,4,opt,name=data,proto3" json:"data,omitempty"`
}

func (x *InstallPluginRequest) GetPluginId() string {
	if x != nil {
		return x.PluginId
	}
	return ""
}
func (x *InstallPluginRequest) GetSource() string {
	if x != nil {
		return x.Source
	}
	return ""
}
func (x *InstallPluginRequest) GetUrl() string {
	if x != nil {
		return x.Url
	}
	return ""
}
func (x *InstallPluginRequest) GetData() []byte {
	if x != nil {
		return x.Data
	}
	return nil
}

// PluginList message
type PluginList struct {
	Plugins []*PluginInfo `protobuf:"bytes,1,rep,name=plugins,proto3" json:"plugins,omitempty"`
}

func (x *PluginList) GetPlugins() []*PluginInfo {
	if x != nil {
		return x.Plugins
	}
	return nil
}

// PluginInfo message
type PluginInfo struct {
	Id          string      `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	Name        string      `protobuf:"bytes,2,opt,name=name,proto3" json:"name,omitempty"`
	Version     string      `protobuf:"bytes,3,opt,name=version,proto3" json:"version,omitempty"`
	Description string      `protobuf:"bytes,4,opt,name=description,proto3" json:"description,omitempty"`
	Author      string      `protobuf:"bytes,5,opt,name=author,proto3" json:"author,omitempty"`
	Icon        string      `protobuf:"bytes,6,opt,name=icon,proto3" json:"icon,omitempty"`
	State       PluginState `protobuf:"varint,7,opt,name=state,proto3,enum=runixo.PluginState" json:"state,omitempty"`
	Type        PluginType  `protobuf:"varint,8,opt,name=type,proto3,enum=runixo.PluginType" json:"type,omitempty"`
	Permissions []string    `protobuf:"bytes,9,rep,name=permissions,proto3" json:"permissions,omitempty"`
	InstalledAt int64       `protobuf:"varint,10,opt,name=installed_at,json=installedAt,proto3" json:"installed_at,omitempty"`
	UpdatedAt   int64       `protobuf:"varint,11,opt,name=updated_at,json=updatedAt,proto3" json:"updated_at,omitempty"`
}

// PluginConfig message
type PluginConfig struct {
	PluginId   string `protobuf:"bytes,1,opt,name=plugin_id,json=pluginId,proto3" json:"plugin_id,omitempty"`
	ConfigJson string `protobuf:"bytes,2,opt,name=config_json,json=configJson,proto3" json:"config_json,omitempty"`
}

func (x *PluginConfig) GetPluginId() string {
	if x != nil {
		return x.PluginId
	}
	return ""
}
func (x *PluginConfig) GetConfigJson() string {
	if x != nil {
		return x.ConfigJson
	}
	return ""
}

// SetPluginConfigRequest message
type SetPluginConfigRequest struct {
	PluginId   string `protobuf:"bytes,1,opt,name=plugin_id,json=pluginId,proto3" json:"plugin_id,omitempty"`
	ConfigJson string `protobuf:"bytes,2,opt,name=config_json,json=configJson,proto3" json:"config_json,omitempty"`
}

func (x *SetPluginConfigRequest) GetPluginId() string {
	if x != nil {
		return x.PluginId
	}
	return ""
}
func (x *SetPluginConfigRequest) GetConfigJson() string {
	if x != nil {
		return x.ConfigJson
	}
	return ""
}

// PluginStatus message
type PluginStatusMsg struct {
	PluginId string            `protobuf:"bytes,1,opt,name=plugin_id,json=pluginId,proto3" json:"plugin_id,omitempty"`
	State    PluginState       `protobuf:"varint,2,opt,name=state,proto3,enum=runixo.PluginState" json:"state,omitempty"`
	Running  bool              `protobuf:"varint,3,opt,name=running,proto3" json:"running,omitempty"`
	Error    string            `protobuf:"bytes,4,opt,name=error,proto3" json:"error,omitempty"`
	Uptime   int64             `protobuf:"varint,5,opt,name=uptime,proto3" json:"uptime,omitempty"`
	Stats    map[string]string `protobuf:"bytes,6,rep,name=stats,proto3" json:"stats,omitempty"`
}

// AvailablePluginList message
type AvailablePluginList struct {
	Plugins []*AvailablePlugin `protobuf:"bytes,1,rep,name=plugins,proto3" json:"plugins,omitempty"`
}

// AvailablePlugin message
type AvailablePlugin struct {
	Id          string     `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	Name        string     `protobuf:"bytes,2,opt,name=name,proto3" json:"name,omitempty"`
	Version     string     `protobuf:"bytes,3,opt,name=version,proto3" json:"version,omitempty"`
	Description string     `protobuf:"bytes,4,opt,name=description,proto3" json:"description,omitempty"`
	Author      string     `protobuf:"bytes,5,opt,name=author,proto3" json:"author,omitempty"`
	Icon        string     `protobuf:"bytes,6,opt,name=icon,proto3" json:"icon,omitempty"`
	Type        PluginType `protobuf:"varint,7,opt,name=type,proto3,enum=runixo.PluginType" json:"type,omitempty"`
	Downloads   int64      `protobuf:"varint,8,opt,name=downloads,proto3" json:"downloads,omitempty"`
	Rating      float64    `protobuf:"fixed64,9,opt,name=rating,proto3" json:"rating,omitempty"`
	RatingCount int32      `protobuf:"varint,10,opt,name=rating_count,json=ratingCount,proto3" json:"rating_count,omitempty"`
	Tags        []string   `protobuf:"bytes,11,rep,name=tags,proto3" json:"tags,omitempty"`
	Category    string     `protobuf:"bytes,12,opt,name=category,proto3" json:"category,omitempty"`
	Official    bool       `protobuf:"varint,13,opt,name=official,proto3" json:"official,omitempty"`
	DownloadUrl string     `protobuf:"bytes,14,opt,name=download_url,json=downloadUrl,proto3" json:"download_url,omitempty"`
	UpdatedAt   string     `protobuf:"bytes,15,opt,name=updated_at,json=updatedAt,proto3" json:"updated_at,omitempty"`
}

// ==================== Update System Types ====================

// UpdateInfo message
type UpdateInfo struct {
	Available      bool   `protobuf:"varint,1,opt,name=available,proto3" json:"available,omitempty"`
	CurrentVersion string `protobuf:"bytes,2,opt,name=current_version,json=currentVersion,proto3" json:"current_version,omitempty"`
	LatestVersion  string `protobuf:"bytes,3,opt,name=latest_version,json=latestVersion,proto3" json:"latest_version,omitempty"`
	ReleaseNotes   string `protobuf:"bytes,4,opt,name=release_notes,json=releaseNotes,proto3" json:"release_notes,omitempty"`
	DownloadUrl    string `protobuf:"bytes,5,opt,name=download_url,json=downloadUrl,proto3" json:"download_url,omitempty"`
	Size           int64  `protobuf:"varint,6,opt,name=size,proto3" json:"size,omitempty"`
	Checksum       string `protobuf:"bytes,7,opt,name=checksum,proto3" json:"checksum,omitempty"`
	ReleaseDate    string `protobuf:"bytes,8,opt,name=release_date,json=releaseDate,proto3" json:"release_date,omitempty"`
	IsCritical     bool   `protobuf:"varint,9,opt,name=is_critical,json=isCritical,proto3" json:"is_critical,omitempty"`
}

// UpdateRequest message
type UpdateRequest struct {
	Version string `protobuf:"bytes,1,opt,name=version,proto3" json:"version,omitempty"`
}

func (x *UpdateRequest) GetVersion() string {
	if x != nil {
		return x.Version
	}
	return ""
}

// DownloadProgress message
type DownloadProgress struct {
	Downloaded int64  `protobuf:"varint,1,opt,name=downloaded,proto3" json:"downloaded,omitempty"`
	Total      int64  `protobuf:"varint,2,opt,name=total,proto3" json:"total,omitempty"`
	Percent    int32  `protobuf:"varint,3,opt,name=percent,proto3" json:"percent,omitempty"`
	Status     string `protobuf:"bytes,4,opt,name=status,proto3" json:"status,omitempty"`
}

// UpdateConfig message
type UpdateConfig struct {
	AutoUpdate    bool   `protobuf:"varint,1,opt,name=auto_update,json=autoUpdate,proto3" json:"auto_update,omitempty"`
	CheckInterval int32  `protobuf:"varint,2,opt,name=check_interval,json=checkInterval,proto3" json:"check_interval,omitempty"`
	UpdateChannel string `protobuf:"bytes,3,opt,name=update_channel,json=updateChannel,proto3" json:"update_channel,omitempty"`
	LastCheck     string `protobuf:"bytes,4,opt,name=last_check,json=lastCheck,proto3" json:"last_check,omitempty"`
	NotifyOnly    bool   `protobuf:"varint,5,opt,name=notify_only,json=notifyOnly,proto3" json:"notify_only,omitempty"`
}

// UpdateHistory message
type UpdateHistory struct {
	Records []*UpdateRecord `protobuf:"bytes,1,rep,name=records,proto3" json:"records,omitempty"`
}

// UpdateRecord message
type UpdateRecord struct {
	Version     string `protobuf:"bytes,1,opt,name=version,proto3" json:"version,omitempty"`
	FromVersion string `protobuf:"bytes,2,opt,name=from_version,json=fromVersion,proto3" json:"from_version,omitempty"`
	Timestamp   int64  `protobuf:"varint,3,opt,name=timestamp,proto3" json:"timestamp,omitempty"`
	Success     bool   `protobuf:"varint,4,opt,name=success,proto3" json:"success,omitempty"`
	Error       string `protobuf:"bytes,5,opt,name=error,proto3" json:"error,omitempty"`
}
