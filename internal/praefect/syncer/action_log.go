package syncer

import (
	"context"

	"github.com/sirupsen/logrus"
	"gitlab.com/gitlab-org/gitaly/v14/internal/praefect/datastore"
)

// LogWarnAction is an implementation of the Action interface that allows to log a warning message
// for the repositories that are not known for the praefect.
type LogWarnAction struct {
	logger logrus.FieldLogger
}

// NewLogWarnAction return new instance of the LogWarnAction.
func NewLogWarnAction(logger logrus.FieldLogger) *LogWarnAction {
	return &LogWarnAction{
		logger: logger.WithField("component", "syncer.log_warn_action"),
	}
}

// Perform for each repository that is not known to praefect this methods log a warning.
func (al LogWarnAction) Perform(_ context.Context, existence map[datastore.RepositoryClusterPath]bool) error {
	for entry, exists := range existence {
		if !exists {
			al.logger.WithFields(logrus.Fields{
				"virtual_storage": entry.VirtualStorage,
				"storage":         entry.Storage,
				"relative_path":   entry.RelativePath,
			}).Warn("repository is not managed by praefect")
		}
	}
	return nil
}
