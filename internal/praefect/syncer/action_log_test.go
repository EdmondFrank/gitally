package syncer

import (
	"context"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/require"
	"gitlab.com/gitlab-org/gitaly/v14/internal/praefect/datastore"
)

func TestLogWarnAction_Perform(t *testing.T) {
	logger, hook := test.NewNullLogger()
	action := NewLogWarnAction(logger)
	err := action.Perform(context.TODO(), map[datastore.RepositoryClusterPath]bool{
		{ClusterPath: datastore.ClusterPath{VirtualStorage: "vs1", Storage: "g1"}, RelativePath: "p/1"}: false,
		{ClusterPath: datastore.ClusterPath{VirtualStorage: "vs2", Storage: "g2"}, RelativePath: "p/2"}: false,
		{ClusterPath: datastore.ClusterPath{VirtualStorage: "vs3", Storage: "g3"}, RelativePath: "p/3"}: true,
	})
	require.NoError(t, err)
	require.Len(t, hook.AllEntries(), 2)

	exp := []map[string]interface{}{{
		"Data": logrus.Fields{
			"component":       "syncer.log_warn_action",
			"virtual_storage": "vs1",
			"storage":         "g1",
			"relative_path":   "p/1",
		},
		"Message": "repository is not managed by praefect",
	}, {
		"Data": logrus.Fields{
			"component":       "syncer.log_warn_action",
			"virtual_storage": "vs2",
			"storage":         "g2",
			"relative_path":   "p/2",
		},
		"Message": "repository is not managed by praefect",
	}}

	require.ElementsMatch(t, exp, []map[string]interface{}{{
		"Data":    hook.AllEntries()[0].Data,
		"Message": hook.AllEntries()[0].Message,
	}, {
		"Data":    hook.AllEntries()[1].Data,
		"Message": hook.AllEntries()[1].Message,
	}})
}
