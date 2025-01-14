package remote

import (
	"context"

	apimodels "github.com/grafana/grafana/pkg/services/ngalert/api/tooling/definitions"
	"github.com/grafana/grafana/pkg/services/ngalert/models"
	"github.com/grafana/grafana/pkg/services/ngalert/notifier"
)

type RemoteSecondaryForkedAlertmanager struct {
	internal notifier.Alertmanager
	remote   notifier.Alertmanager
}

func NewRemoteSecondaryForkedAlertmanager(internal, remote notifier.Alertmanager) *RemoteSecondaryForkedAlertmanager {
	return &RemoteSecondaryForkedAlertmanager{
		internal: internal,
		remote:   remote,
	}
}

func (fam *RemoteSecondaryForkedAlertmanager) ApplyConfig(ctx context.Context, config *models.AlertConfiguration) error {
	return nil
}

func (fam *RemoteSecondaryForkedAlertmanager) SaveAndApplyConfig(ctx context.Context, config *apimodels.PostableUserConfig) error {
	return nil
}

func (fam *RemoteSecondaryForkedAlertmanager) SaveAndApplyDefaultConfig(ctx context.Context) error {
	return nil
}

func (fam *RemoteSecondaryForkedAlertmanager) GetStatus() apimodels.GettableStatus {
	return apimodels.GettableStatus{}
}

func (fam *RemoteSecondaryForkedAlertmanager) CreateSilence(ctx context.Context, silence *apimodels.PostableSilence) (string, error) {
	return "", nil
}

func (fam *RemoteSecondaryForkedAlertmanager) DeleteSilence(ctx context.Context, id string) error {
	return nil
}

func (fam *RemoteSecondaryForkedAlertmanager) GetSilence(ctx context.Context, id string) (apimodels.GettableSilence, error) {
	return apimodels.GettableSilence{}, nil
}

func (fam *RemoteSecondaryForkedAlertmanager) ListSilences(ctx context.Context, filter []string) (apimodels.GettableSilences, error) {
	return apimodels.GettableSilences{}, nil
}

func (fam *RemoteSecondaryForkedAlertmanager) GetAlerts(ctx context.Context, active, silenced, inhibited bool, filter []string, receiver string) (apimodels.GettableAlerts, error) {
	return apimodels.GettableAlerts{}, nil
}

func (fam *RemoteSecondaryForkedAlertmanager) GetAlertGroups(ctx context.Context, active, silenced, inhibited bool, filter []string, receiver string) (apimodels.AlertGroups, error) {
	return apimodels.AlertGroups{}, nil
}

func (fam *RemoteSecondaryForkedAlertmanager) PutAlerts(ctx context.Context, alerts apimodels.PostableAlerts) error {
	return nil
}

func (fam *RemoteSecondaryForkedAlertmanager) GetReceivers(ctx context.Context) ([]apimodels.Receiver, error) {
	return []apimodels.Receiver{}, nil
}

func (fam *RemoteSecondaryForkedAlertmanager) TestReceivers(ctx context.Context, c apimodels.TestReceiversConfigBodyParams) (*notifier.TestReceiversResult, error) {
	return &notifier.TestReceiversResult{}, nil
}

func (fam *RemoteSecondaryForkedAlertmanager) TestTemplate(ctx context.Context, c apimodels.TestTemplatesConfigBodyParams) (*notifier.TestTemplatesResults, error) {
	return &notifier.TestTemplatesResults{}, nil
}

func (fam *RemoteSecondaryForkedAlertmanager) CleanUp() {}

func (fam *RemoteSecondaryForkedAlertmanager) StopAndWait() {}

func (fam *RemoteSecondaryForkedAlertmanager) Ready() bool {
	return false
}
