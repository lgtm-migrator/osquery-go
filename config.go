package osquery

import (
	"context"

	"github.com/kolide/osquery-golang/gen/osquery"
)

// ConfigPlugin is the minimum interface required to implement an osquery
// config plugin. Any value that implements this interface can be passed to
// NewConfigPlugin to satisfy the full OsqueryPlugin interface.
type ConfigPlugin interface {
	// Name returns the name of the config plugin.
	Name() string

	// GenerateConfigs returns the configurations generated by this plugin.
	// The returned map should use the source name as key, and the config
	// JSON as values. The context argument can optionally be used for
	// cancellation in long-running operations.
	GenerateConfigs(ctx context.Context) (map[string]string, error)
}

// NewConfigPlugin takes a value that implements ConfigPlugin and wraps it with
// the appropriate methods to satisfy the OsqueryPlugin interface. Use this to
// easily create plugins implementing osquery tables.
func NewConfigPlugin(plugin ConfigPlugin) *configPluginImpl {
	return &configPluginImpl{plugin}
}

type configPluginImpl struct {
	plugin ConfigPlugin
}

func (t *configPluginImpl) Name() string {
	return t.plugin.Name()
}

// Registry name for config plugins
const configRegistryName = "config"

func (t *configPluginImpl) RegistryName() string {
	return configRegistryName
}

func (t *configPluginImpl) Routes() osquery.ExtensionPluginResponse {
	return osquery.ExtensionPluginResponse{}
}

func (t *configPluginImpl) Ping() osquery.ExtensionStatus {
	return StatusOK
}

// Key that the request method is stored under
const requestActionKey = "action"

// Action value used when config is requested
const genConfigAction = "genConfig"

func (t *configPluginImpl) Call(ctx context.Context, request osquery.ExtensionPluginRequest) osquery.ExtensionResponse {
	switch request[requestActionKey] {
	case genConfigAction:
		configs, err := t.plugin.GenerateConfigs(ctx)
		if err != nil {
			return osquery.ExtensionResponse{
				Status: &osquery.ExtensionStatus{
					Code:    1,
					Message: "error getting config: " + err.Error(),
				},
			}
		}

		return osquery.ExtensionResponse{
			Status:   &StatusOK,
			Response: osquery.ExtensionPluginResponse{configs},
		}

	default:
		return osquery.ExtensionResponse{
			Status: &osquery.ExtensionStatus{
				Code:    1,
				Message: "unknown action: " + request["action"],
			},
		}
	}

}

func (t *configPluginImpl) Shutdown() {}