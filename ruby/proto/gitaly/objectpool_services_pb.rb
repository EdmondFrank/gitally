# Generated by the protocol buffer compiler.  DO NOT EDIT!
# Source: objectpool.proto for package 'gitaly'

require 'grpc'
require 'objectpool_pb'

module Gitaly
  module ObjectPoolService
    class Service

      include GRPC::GenericService

      self.marshal_class_method = :encode
      self.unmarshal_class_method = :decode
      self.service_name = 'gitaly.ObjectPoolService'

      rpc :CreateObjectPool, Gitaly::CreateObjectPoolRequest, Gitaly::CreateObjectPoolResponse
      rpc :DeleteObjectPool, Gitaly::DeleteObjectPoolRequest, Gitaly::DeleteObjectPoolResponse
      # Repositories are assumed to be stored on the same disk
      rpc :LinkRepositoryToObjectPool, Gitaly::LinkRepositoryToObjectPoolRequest, Gitaly::LinkRepositoryToObjectPoolResponse
      # UnlinkRepositoryFromObjectPool does not unlink the repository from the
      # object pool as you'd think, but all it really does is to remove the object
      # pool's remote pointing to the repository. And even this is a no-op given
      # that we'd try to remove the remote by the repository's `GlRepository()`
      # name, which we never create in the first place. To unlink repositories
      # from an object pool, you'd really want to execute DisconnectGitAlternates
      # to remove the repository's link to the pool's object database.
      #
      # This function is never called by anyone and highly misleading. It's thus
      # deprecated and will be removed in v14.4.
      rpc :UnlinkRepositoryFromObjectPool, Gitaly::UnlinkRepositoryFromObjectPoolRequest, Gitaly::UnlinkRepositoryFromObjectPoolResponse
      rpc :ReduplicateRepository, Gitaly::ReduplicateRepositoryRequest, Gitaly::ReduplicateRepositoryResponse
      rpc :DisconnectGitAlternates, Gitaly::DisconnectGitAlternatesRequest, Gitaly::DisconnectGitAlternatesResponse
      rpc :FetchIntoObjectPool, Gitaly::FetchIntoObjectPoolRequest, Gitaly::FetchIntoObjectPoolResponse
      rpc :GetObjectPool, Gitaly::GetObjectPoolRequest, Gitaly::GetObjectPoolResponse
    end

    Stub = Service.rpc_stub_class
  end
end
