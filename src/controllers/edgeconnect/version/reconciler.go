package version

import (
	"context"

	edgeconnectv1alpha1 "github.com/Dynatrace/dynatrace-operator/src/api/v1alpha1/edgeconnect"
	"github.com/Dynatrace/dynatrace-operator/src/timeprovider"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Reconciler struct {
	edgeConnect  *edgeconnectv1alpha1.EdgeConnect
	timeProvider *timeprovider.Provider

	apiReader client.Reader
}

func NewReconciler(edgeConnect *edgeconnectv1alpha1.EdgeConnect, apiReader client.Reader, timeProvider *timeprovider.Provider) *Reconciler { //nolint:revive
	return &Reconciler{
		edgeConnect:  edgeConnect,
		apiReader:    apiReader,
		timeProvider: timeProvider,
	}
}

func (reconciler *Reconciler) Reconcile(ctx context.Context) error {
	updaters := []versionStatusUpdater{
		newEdgeConnectUpdater(reconciler.edgeConnect, reconciler.apiReader, reconciler.timeProvider),
	}

	for _, updater := range updaters {
		log.Info("updating version status", "updater", updater.Name())

		if updater.RequiresReconcile() {
			return updater.Update(ctx)
		}

		log.Info("no reconcile required", "updater", updater.Name())
	}

	return nil
}
