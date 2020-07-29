package info

import (
	"context"
	"sort"

	"gitlab.com/gitlab-org/gitaly/proto/go/gitalypb"
)

func (s *Server) DatalossCheck(ctx context.Context, req *gitalypb.DatalossCheckRequest) (*gitalypb.DatalossCheckResponse, error) {
	shard, err := s.nodeMgr.GetShard(req.GetVirtualStorage())
	if err != nil {
		return nil, err
	}

	outdatedRepos, err := s.rs.GetOutdatedRepositories(ctx, req.GetVirtualStorage())
	if err != nil {
		return nil, err
	}

	pbRepos := make([]*gitalypb.DatalossCheckResponse_Repository, 0, len(outdatedRepos))
	for relativePath, storages := range outdatedRepos {
		pbStorages := make([]*gitalypb.DatalossCheckResponse_Repository_Storage, 0, len(storages))
		for name, behindBy := range storages {
			pbStorages = append(pbStorages, &gitalypb.DatalossCheckResponse_Repository_Storage{
				Name:     name,
				BehindBy: int64(behindBy),
			})
		}

		sort.Slice(pbStorages, func(i, j int) bool { return pbStorages[i].Name < pbStorages[j].Name })

		pbRepos = append(pbRepos, &gitalypb.DatalossCheckResponse_Repository{
			RelativePath: relativePath,
			Storages:     pbStorages,
		})
	}

	sort.Slice(pbRepos, func(i, j int) bool { return pbRepos[i].RelativePath < pbRepos[j].RelativePath })

	return &gitalypb.DatalossCheckResponse{
		Primary:      shard.Primary.GetStorage(),
		Repositories: pbRepos,
	}, nil
}
