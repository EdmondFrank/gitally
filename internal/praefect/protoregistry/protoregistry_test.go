package protoregistry_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"gitlab.com/gitlab-org/gitaly/v14/internal/praefect/protoregistry"
	"gitlab.com/gitlab-org/gitaly/v14/internal/testhelper/testassert"
	"gitlab.com/gitlab-org/gitaly/v14/proto/go/gitalypb"
)

func TestNewProtoRegistry(t *testing.T) {
	expectedResults := map[string]map[string]protoregistry.OpType{
		"BlobService": {
			"GetBlob":        protoregistry.OpAccessor,
			"GetBlobs":       protoregistry.OpAccessor,
			"GetLFSPointers": protoregistry.OpAccessor,
		},
		"CleanupService": {
			"ApplyBfgObjectMapStream": protoregistry.OpMutator,
		},
		"CommitService": {
			"CommitIsAncestor":         protoregistry.OpAccessor,
			"TreeEntry":                protoregistry.OpAccessor,
			"CommitsBetween":           protoregistry.OpAccessor,
			"CountCommits":             protoregistry.OpAccessor,
			"CountDivergingCommits":    protoregistry.OpAccessor,
			"GetTreeEntries":           protoregistry.OpAccessor,
			"ListFiles":                protoregistry.OpAccessor,
			"FindCommit":               protoregistry.OpAccessor,
			"CommitStats":              protoregistry.OpAccessor,
			"FindAllCommits":           protoregistry.OpAccessor,
			"FindCommits":              protoregistry.OpAccessor,
			"CommitLanguages":          protoregistry.OpAccessor,
			"RawBlame":                 protoregistry.OpAccessor,
			"LastCommitForPath":        protoregistry.OpAccessor,
			"ListLastCommitsForTree":   protoregistry.OpAccessor,
			"CommitsByMessage":         protoregistry.OpAccessor,
			"ListCommitsByOid":         protoregistry.OpAccessor,
			"FilterShasWithSignatures": protoregistry.OpAccessor,
		},
		"ConflictsService": {
			"ListConflictFiles": protoregistry.OpAccessor,
			"ResolveConflicts":  protoregistry.OpMutator,
		},
		"DiffService": {
			"CommitDiff":  protoregistry.OpAccessor,
			"CommitDelta": protoregistry.OpAccessor,
			"RawDiff":     protoregistry.OpAccessor,
			"RawPatch":    protoregistry.OpAccessor,
			"DiffStats":   protoregistry.OpAccessor,
		},
		"NamespaceService": {
			"AddNamespace":    protoregistry.OpMutator,
			"RemoveNamespace": protoregistry.OpMutator,
			"RenameNamespace": protoregistry.OpMutator,
			"NamespaceExists": protoregistry.OpAccessor,
		},
		"ObjectPoolService": {
			"CreateObjectPool":               protoregistry.OpMutator,
			"DeleteObjectPool":               protoregistry.OpMutator,
			"LinkRepositoryToObjectPool":     protoregistry.OpMutator,
			"UnlinkRepositoryFromObjectPool": protoregistry.OpMutator,
			"ReduplicateRepository":          protoregistry.OpMutator,
			"DisconnectGitAlternates":        protoregistry.OpMutator,
		},
		"OperationService": {
			"UserCreateBranch":    protoregistry.OpMutator,
			"UserUpdateBranch":    protoregistry.OpMutator,
			"UserDeleteBranch":    protoregistry.OpMutator,
			"UserCreateTag":       protoregistry.OpMutator,
			"UserDeleteTag":       protoregistry.OpMutator,
			"UserMergeToRef":      protoregistry.OpMutator,
			"UserMergeBranch":     protoregistry.OpMutator,
			"UserFFBranch":        protoregistry.OpMutator,
			"UserCherryPick":      protoregistry.OpMutator,
			"UserCommitFiles":     protoregistry.OpMutator,
			"UserRevert":          protoregistry.OpMutator,
			"UserSquash":          protoregistry.OpMutator,
			"UserApplyPatch":      protoregistry.OpMutator,
			"UserUpdateSubmodule": protoregistry.OpMutator,
		},
		"RefService": {
			"FindDefaultBranchName":           protoregistry.OpAccessor,
			"FindAllBranchNames":              protoregistry.OpAccessor,
			"FindAllTagNames":                 protoregistry.OpAccessor,
			"FindLocalBranches":               protoregistry.OpAccessor,
			"FindAllBranches":                 protoregistry.OpAccessor,
			"FindAllTags":                     protoregistry.OpAccessor,
			"FindAllRemoteBranches":           protoregistry.OpAccessor,
			"RefExists":                       protoregistry.OpAccessor,
			"FindBranch":                      protoregistry.OpAccessor,
			"DeleteRefs":                      protoregistry.OpMutator,
			"ListBranchNamesContainingCommit": protoregistry.OpAccessor,
			"ListTagNamesContainingCommit":    protoregistry.OpAccessor,
			"GetTagMessages":                  protoregistry.OpAccessor,
			"ListNewCommits":                  protoregistry.OpAccessor,
			"ListNewBlobs":                    protoregistry.OpAccessor,
			"PackRefs":                        protoregistry.OpMutator,
		},
		"RemoteService": {
			"FetchInternalRemote":  protoregistry.OpMutator,
			"UpdateRemoteMirror":   protoregistry.OpAccessor,
			"FindRemoteRepository": protoregistry.OpAccessor,
			"FindRemoteRootRef":    protoregistry.OpAccessor,
		},
		"RepositoryService": {
			"RepackIncremental":            protoregistry.OpMutator,
			"RepackFull":                   protoregistry.OpMutator,
			"GarbageCollect":               protoregistry.OpMutator,
			"RepositorySize":               protoregistry.OpAccessor,
			"ApplyGitattributes":           protoregistry.OpMutator,
			"FetchRemote":                  protoregistry.OpMutator,
			"CreateRepository":             protoregistry.OpMutator,
			"GetArchive":                   protoregistry.OpAccessor,
			"HasLocalBranches":             protoregistry.OpAccessor,
			"FetchSourceBranch":            protoregistry.OpMutator,
			"Fsck":                         protoregistry.OpAccessor,
			"WriteRef":                     protoregistry.OpMutator,
			"FindMergeBase":                protoregistry.OpAccessor,
			"CreateFork":                   protoregistry.OpMutator,
			"IsSquashInProgress":           protoregistry.OpAccessor,
			"CreateRepositoryFromURL":      protoregistry.OpMutator,
			"CreateBundle":                 protoregistry.OpAccessor,
			"CreateRepositoryFromBundle":   protoregistry.OpMutator,
			"SetConfig":                    protoregistry.OpMutator,
			"DeleteConfig":                 protoregistry.OpMutator,
			"FindLicense":                  protoregistry.OpAccessor,
			"GetInfoAttributes":            protoregistry.OpAccessor,
			"CalculateChecksum":            protoregistry.OpAccessor,
			"Cleanup":                      protoregistry.OpMutator,
			"GetSnapshot":                  protoregistry.OpAccessor,
			"CreateRepositoryFromSnapshot": protoregistry.OpMutator,
			"GetRawChanges":                protoregistry.OpAccessor,
			"SearchFilesByContent":         protoregistry.OpAccessor,
			"SearchFilesByName":            protoregistry.OpAccessor,
			"RestoreCustomHooks":           protoregistry.OpMutator,
			"BackupCustomHooks":            protoregistry.OpAccessor,
		},
		"SmartHTTPService": {
			"InfoRefsUploadPack":  protoregistry.OpAccessor,
			"InfoRefsReceivePack": protoregistry.OpAccessor,
			"PostUploadPack":      protoregistry.OpAccessor,
			"PostReceivePack":     protoregistry.OpMutator,
		},
		"SSHService": {
			"SSHUploadPack":    protoregistry.OpAccessor,
			"SSHReceivePack":   protoregistry.OpMutator,
			"SSHUploadArchive": protoregistry.OpAccessor,
		},
		"WikiService": {
			"WikiWritePage":   protoregistry.OpMutator,
			"WikiUpdatePage":  protoregistry.OpMutator,
			"WikiFindPage":    protoregistry.OpAccessor,
			"WikiGetAllPages": protoregistry.OpAccessor,
			"WikiListPages":   protoregistry.OpAccessor,
		},
	}

	for serviceName, methods := range expectedResults {
		for methodName, opType := range methods {
			method := fmt.Sprintf("/gitaly.%s/%s", serviceName, methodName)

			methodInfo, err := protoregistry.GitalyProtoPreregistered.LookupMethod(method)
			require.NoError(t, err)

			require.Equalf(t, opType, methodInfo.Operation, "expect %s:%s to have the correct op type", serviceName, methodName)
			require.Equal(t, method, methodInfo.FullMethodName())
			require.False(t, protoregistry.GitalyProtoPreregistered.IsInterceptedMethod(method), method)
		}
	}
}

func TestNewProtoRegistry_IsInterceptedMethod(t *testing.T) {
	for service, methods := range map[string][]string{
		"ServerService": {
			"ServerInfo",
			"DiskStatistics",
		},
		"PraefectInfoService": {
			"RepositoryReplicas",
			"DatalossCheck",
			"SetAuthoritativeStorage",
		},
	} {
		t.Run(service, func(t *testing.T) {
			for _, method := range methods {
				t.Run(method, func(t *testing.T) {
					fullMethodName := fmt.Sprintf("/gitaly.%s/%s", service, method)
					require.True(t, protoregistry.GitalyProtoPreregistered.IsInterceptedMethod(fullMethodName))
					methodInfo, err := protoregistry.GitalyProtoPreregistered.LookupMethod(fullMethodName)
					require.Empty(t, methodInfo)
					require.Error(t, err, "full method name not found:")
				})
			}
		})
	}
}

func TestRequestFactory(t *testing.T) {
	mInfo, err := protoregistry.GitalyProtoPreregistered.LookupMethod("/gitaly.RepositoryService/GarbageCollect")
	require.NoError(t, err)

	pb, err := mInfo.UnmarshalRequestProto([]byte{})
	require.NoError(t, err)

	testassert.ProtoEqual(t, &gitalypb.GarbageCollectRequest{}, pb)
}

func TestMethodInfoScope(t *testing.T) {
	for _, tt := range []struct {
		method string
		scope  protoregistry.Scope
	}{
		{
			method: "/gitaly.RepositoryService/GarbageCollect",
			scope:  protoregistry.ScopeRepository,
		},
	} {
		t.Run(tt.method, func(t *testing.T) {
			mInfo, err := protoregistry.GitalyProtoPreregistered.LookupMethod(tt.method)
			require.NoError(t, err)

			require.Exactly(t, tt.scope, mInfo.Scope)
		})
	}
}
