package csigc

import (
	"context"
	"os"

	"github.com/spf13/afero"
)

//nolint:gosec
func (gc *CSIGarbageCollector) runBinaryGarbageCollection(ctx context.Context) {
	fs := &afero.Afero{Fs: gc.fs}

	gcRunsMetric.Inc()

	codeModules, err := gc.db.ReadCodeModules(ctx)
	if err != nil {
		log.Error(err, "failed to read codemodules")

		return
	}

	for _, codeModule := range codeModules {
		orphaned, err := gc.db.IsCodeModuleOrphaned(ctx, &codeModule)
		if err != nil {
			log.Error(err, "failed to check if codemodule is orphaned")

			continue
		}

		if orphaned {
			removeUnusedVersion(fs, codeModule.Location)

			err := gc.db.DeleteCodeModule(ctx, &codeModule)
			if err != nil {
				log.Error(err, "failed to delete codemodule")

				return
			}
		}
	}
}

func removeUnusedVersion(fs *afero.Afero, binaryPath string) {
	size, _ := dirSize(fs, binaryPath)

	err := fs.RemoveAll(binaryPath)
	if err != nil {
		log.Info("delete failed", "path", binaryPath)
	} else {
		foldersRemovedMetric.Inc()
		reclaimedMemoryMetric.Add(float64(size))
	}
}

func dirSize(fs *afero.Afero, path string) (int64, error) {
	var size int64

	err := fs.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			size += info.Size()
		}

		return err
	})

	return size, err
}
