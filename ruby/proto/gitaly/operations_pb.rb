# Generated by the protocol buffer compiler.  DO NOT EDIT!
# source: operations.proto

require 'google/protobuf'

require 'lint_pb'
require 'shared_pb'
require 'google/protobuf/timestamp_pb'
Google::Protobuf::DescriptorPool.generated_pool.build do
  add_file("operations.proto", :syntax => :proto3) do
    add_message "gitaly.UserCreateBranchRequest" do
      optional :repository, :message, 1, "gitaly.Repository"
      optional :branch_name, :bytes, 2
      optional :user, :message, 3, "gitaly.User"
      optional :start_point, :bytes, 4
    end
    add_message "gitaly.UserCreateBranchResponse" do
      optional :branch, :message, 1, "gitaly.Branch"
      optional :pre_receive_error, :string, 2
    end
    add_message "gitaly.UserUpdateBranchRequest" do
      optional :repository, :message, 1, "gitaly.Repository"
      optional :branch_name, :bytes, 2
      optional :user, :message, 3, "gitaly.User"
      optional :newrev, :bytes, 4
      optional :oldrev, :bytes, 5
    end
    add_message "gitaly.UserUpdateBranchResponse" do
      optional :pre_receive_error, :string, 1
    end
    add_message "gitaly.UserDeleteBranchRequest" do
      optional :repository, :message, 1, "gitaly.Repository"
      optional :branch_name, :bytes, 2
      optional :user, :message, 3, "gitaly.User"
    end
    add_message "gitaly.UserDeleteBranchResponse" do
      optional :pre_receive_error, :string, 1
    end
    add_message "gitaly.UserDeleteTagRequest" do
      optional :repository, :message, 1, "gitaly.Repository"
      optional :tag_name, :bytes, 2
      optional :user, :message, 3, "gitaly.User"
    end
    add_message "gitaly.UserDeleteTagResponse" do
      optional :pre_receive_error, :string, 1
    end
    add_message "gitaly.UserCreateTagRequest" do
      optional :repository, :message, 1, "gitaly.Repository"
      optional :tag_name, :bytes, 2
      optional :user, :message, 3, "gitaly.User"
      optional :target_revision, :bytes, 4
      optional :message, :bytes, 5
    end
    add_message "gitaly.UserCreateTagResponse" do
      optional :tag, :message, 1, "gitaly.Tag"
      optional :exists, :bool, 2
      optional :pre_receive_error, :string, 3
    end
    add_message "gitaly.UserMergeBranchRequest" do
      optional :repository, :message, 1, "gitaly.Repository"
      optional :user, :message, 2, "gitaly.User"
      optional :commit_id, :string, 3
      optional :branch, :bytes, 4
      optional :message, :bytes, 5
      optional :timestamp, :message, 7, "google.protobuf.Timestamp"
      optional :apply, :bool, 6
    end
    add_message "gitaly.UserMergeBranchResponse" do
      optional :commit_id, :string, 1
      optional :branch_update, :message, 3, "gitaly.OperationBranchUpdate"
      optional :pre_receive_error, :string, 4
    end
    add_message "gitaly.UserMergeToRefRequest" do
      optional :repository, :message, 1, "gitaly.Repository"
      optional :user, :message, 2, "gitaly.User"
      optional :source_sha, :string, 3
      optional :branch, :bytes, 4
      optional :target_ref, :bytes, 5
      optional :message, :bytes, 6
      optional :first_parent_ref, :bytes, 7
      optional :allow_conflicts, :bool, 8
      optional :timestamp, :message, 9, "google.protobuf.Timestamp"
    end
    add_message "gitaly.UserMergeToRefResponse" do
      optional :commit_id, :string, 1
      optional :pre_receive_error, :string, 2
    end
    add_message "gitaly.OperationBranchUpdate" do
      optional :commit_id, :string, 1
      optional :repo_created, :bool, 2
      optional :branch_created, :bool, 3
    end
    add_message "gitaly.UserFFBranchRequest" do
      optional :repository, :message, 1, "gitaly.Repository"
      optional :user, :message, 2, "gitaly.User"
      optional :commit_id, :string, 3
      optional :branch, :bytes, 4
    end
    add_message "gitaly.UserFFBranchResponse" do
      optional :branch_update, :message, 1, "gitaly.OperationBranchUpdate"
      optional :pre_receive_error, :string, 2
    end
    add_message "gitaly.UserCherryPickRequest" do
      optional :repository, :message, 1, "gitaly.Repository"
      optional :user, :message, 2, "gitaly.User"
      optional :commit, :message, 3, "gitaly.GitCommit"
      optional :branch_name, :bytes, 4
      optional :message, :bytes, 5
      optional :start_branch_name, :bytes, 6
      optional :start_repository, :message, 7, "gitaly.Repository"
      optional :dry_run, :bool, 8
    end
    add_message "gitaly.UserCherryPickResponse" do
      optional :branch_update, :message, 1, "gitaly.OperationBranchUpdate"
      optional :create_tree_error, :string, 2
      optional :commit_error, :string, 3
      optional :pre_receive_error, :string, 4
      optional :create_tree_error_code, :enum, 5, "gitaly.UserCherryPickResponse.CreateTreeError"
    end
    add_enum "gitaly.UserCherryPickResponse.CreateTreeError" do
      value :NONE, 0
      value :EMPTY, 1
      value :CONFLICT, 2
    end
    add_message "gitaly.UserRevertRequest" do
      optional :repository, :message, 1, "gitaly.Repository"
      optional :user, :message, 2, "gitaly.User"
      optional :commit, :message, 3, "gitaly.GitCommit"
      optional :branch_name, :bytes, 4
      optional :message, :bytes, 5
      optional :start_branch_name, :bytes, 6
      optional :start_repository, :message, 7, "gitaly.Repository"
      optional :dry_run, :bool, 8
    end
    add_message "gitaly.UserRevertResponse" do
      optional :branch_update, :message, 1, "gitaly.OperationBranchUpdate"
      optional :create_tree_error, :string, 2
      optional :commit_error, :string, 3
      optional :pre_receive_error, :string, 4
      optional :create_tree_error_code, :enum, 5, "gitaly.UserRevertResponse.CreateTreeError"
    end
    add_enum "gitaly.UserRevertResponse.CreateTreeError" do
      value :NONE, 0
      value :EMPTY, 1
      value :CONFLICT, 2
    end
    add_message "gitaly.UserCommitFilesActionHeader" do
      optional :action, :enum, 1, "gitaly.UserCommitFilesActionHeader.ActionType"
      optional :file_path, :bytes, 2
      optional :previous_path, :bytes, 3
      optional :base64_content, :bool, 4
      optional :execute_filemode, :bool, 5
      optional :infer_content, :bool, 6
    end
    add_enum "gitaly.UserCommitFilesActionHeader.ActionType" do
      value :CREATE, 0
      value :CREATE_DIR, 1
      value :UPDATE, 2
      value :MOVE, 3
      value :DELETE, 4
      value :CHMOD, 5
    end
    add_message "gitaly.UserCommitFilesAction" do
      oneof :user_commit_files_action_payload do
        optional :header, :message, 1, "gitaly.UserCommitFilesActionHeader"
        optional :content, :bytes, 2
      end
    end
    add_message "gitaly.UserCommitFilesRequestHeader" do
      optional :repository, :message, 1, "gitaly.Repository"
      optional :user, :message, 2, "gitaly.User"
      optional :branch_name, :bytes, 3
      optional :commit_message, :bytes, 4
      optional :commit_author_name, :bytes, 5
      optional :commit_author_email, :bytes, 6
      optional :start_branch_name, :bytes, 7
      optional :start_repository, :message, 8, "gitaly.Repository"
      optional :force, :bool, 9
      optional :start_sha, :string, 10
    end
    add_message "gitaly.UserCommitFilesRequest" do
      oneof :user_commit_files_request_payload do
        optional :header, :message, 1, "gitaly.UserCommitFilesRequestHeader"
        optional :action, :message, 2, "gitaly.UserCommitFilesAction"
      end
    end
    add_message "gitaly.UserCommitFilesResponse" do
      optional :branch_update, :message, 1, "gitaly.OperationBranchUpdate"
      optional :index_error, :string, 2
      optional :pre_receive_error, :string, 3
    end
    add_message "gitaly.UserRebaseConfirmableRequest" do
      oneof :user_rebase_confirmable_request_payload do
        optional :header, :message, 1, "gitaly.UserRebaseConfirmableRequest.Header"
        optional :apply, :bool, 2
      end
    end
    add_message "gitaly.UserRebaseConfirmableRequest.Header" do
      optional :repository, :message, 1, "gitaly.Repository"
      optional :user, :message, 2, "gitaly.User"
      optional :rebase_id, :string, 3
      optional :branch, :bytes, 4
      optional :branch_sha, :string, 5
      optional :remote_repository, :message, 6, "gitaly.Repository"
      optional :remote_branch, :bytes, 7
      repeated :git_push_options, :string, 8
      optional :timestamp, :message, 9, "google.protobuf.Timestamp"
    end
    add_message "gitaly.UserRebaseConfirmableResponse" do
      optional :pre_receive_error, :string, 3
      optional :git_error, :string, 4
      oneof :user_rebase_confirmable_response_payload do
        optional :rebase_sha, :string, 1
        optional :rebase_applied, :bool, 2
      end
    end
    add_message "gitaly.UserSquashRequest" do
      optional :repository, :message, 1, "gitaly.Repository"
      optional :user, :message, 2, "gitaly.User"
      optional :squash_id, :string, 3
      optional :start_sha, :string, 5
      optional :end_sha, :string, 6
      optional :author, :message, 7, "gitaly.User"
      optional :commit_message, :bytes, 8
    end
    add_message "gitaly.UserSquashResponse" do
      optional :squash_sha, :string, 1
      optional :git_error, :string, 3
    end
    add_message "gitaly.UserApplyPatchRequest" do
      oneof :user_apply_patch_request_payload do
        optional :header, :message, 1, "gitaly.UserApplyPatchRequest.Header"
        optional :patches, :bytes, 2
      end
    end
    add_message "gitaly.UserApplyPatchRequest.Header" do
      optional :repository, :message, 1, "gitaly.Repository"
      optional :user, :message, 2, "gitaly.User"
      optional :target_branch, :bytes, 3
    end
    add_message "gitaly.UserApplyPatchResponse" do
      optional :branch_update, :message, 1, "gitaly.OperationBranchUpdate"
    end
    add_message "gitaly.UserUpdateSubmoduleRequest" do
      optional :repository, :message, 1, "gitaly.Repository"
      optional :user, :message, 2, "gitaly.User"
      optional :commit_sha, :string, 3
      optional :branch, :bytes, 4
      optional :submodule, :bytes, 5
      optional :commit_message, :bytes, 6
    end
    add_message "gitaly.UserUpdateSubmoduleResponse" do
      optional :branch_update, :message, 1, "gitaly.OperationBranchUpdate"
      optional :pre_receive_error, :string, 2
      optional :commit_error, :string, 4
    end
  end
end

module Gitaly
  UserCreateBranchRequest = ::Google::Protobuf::DescriptorPool.generated_pool.lookup("gitaly.UserCreateBranchRequest").msgclass
  UserCreateBranchResponse = ::Google::Protobuf::DescriptorPool.generated_pool.lookup("gitaly.UserCreateBranchResponse").msgclass
  UserUpdateBranchRequest = ::Google::Protobuf::DescriptorPool.generated_pool.lookup("gitaly.UserUpdateBranchRequest").msgclass
  UserUpdateBranchResponse = ::Google::Protobuf::DescriptorPool.generated_pool.lookup("gitaly.UserUpdateBranchResponse").msgclass
  UserDeleteBranchRequest = ::Google::Protobuf::DescriptorPool.generated_pool.lookup("gitaly.UserDeleteBranchRequest").msgclass
  UserDeleteBranchResponse = ::Google::Protobuf::DescriptorPool.generated_pool.lookup("gitaly.UserDeleteBranchResponse").msgclass
  UserDeleteTagRequest = ::Google::Protobuf::DescriptorPool.generated_pool.lookup("gitaly.UserDeleteTagRequest").msgclass
  UserDeleteTagResponse = ::Google::Protobuf::DescriptorPool.generated_pool.lookup("gitaly.UserDeleteTagResponse").msgclass
  UserCreateTagRequest = ::Google::Protobuf::DescriptorPool.generated_pool.lookup("gitaly.UserCreateTagRequest").msgclass
  UserCreateTagResponse = ::Google::Protobuf::DescriptorPool.generated_pool.lookup("gitaly.UserCreateTagResponse").msgclass
  UserMergeBranchRequest = ::Google::Protobuf::DescriptorPool.generated_pool.lookup("gitaly.UserMergeBranchRequest").msgclass
  UserMergeBranchResponse = ::Google::Protobuf::DescriptorPool.generated_pool.lookup("gitaly.UserMergeBranchResponse").msgclass
  UserMergeToRefRequest = ::Google::Protobuf::DescriptorPool.generated_pool.lookup("gitaly.UserMergeToRefRequest").msgclass
  UserMergeToRefResponse = ::Google::Protobuf::DescriptorPool.generated_pool.lookup("gitaly.UserMergeToRefResponse").msgclass
  OperationBranchUpdate = ::Google::Protobuf::DescriptorPool.generated_pool.lookup("gitaly.OperationBranchUpdate").msgclass
  UserFFBranchRequest = ::Google::Protobuf::DescriptorPool.generated_pool.lookup("gitaly.UserFFBranchRequest").msgclass
  UserFFBranchResponse = ::Google::Protobuf::DescriptorPool.generated_pool.lookup("gitaly.UserFFBranchResponse").msgclass
  UserCherryPickRequest = ::Google::Protobuf::DescriptorPool.generated_pool.lookup("gitaly.UserCherryPickRequest").msgclass
  UserCherryPickResponse = ::Google::Protobuf::DescriptorPool.generated_pool.lookup("gitaly.UserCherryPickResponse").msgclass
  UserCherryPickResponse::CreateTreeError = ::Google::Protobuf::DescriptorPool.generated_pool.lookup("gitaly.UserCherryPickResponse.CreateTreeError").enummodule
  UserRevertRequest = ::Google::Protobuf::DescriptorPool.generated_pool.lookup("gitaly.UserRevertRequest").msgclass
  UserRevertResponse = ::Google::Protobuf::DescriptorPool.generated_pool.lookup("gitaly.UserRevertResponse").msgclass
  UserRevertResponse::CreateTreeError = ::Google::Protobuf::DescriptorPool.generated_pool.lookup("gitaly.UserRevertResponse.CreateTreeError").enummodule
  UserCommitFilesActionHeader = ::Google::Protobuf::DescriptorPool.generated_pool.lookup("gitaly.UserCommitFilesActionHeader").msgclass
  UserCommitFilesActionHeader::ActionType = ::Google::Protobuf::DescriptorPool.generated_pool.lookup("gitaly.UserCommitFilesActionHeader.ActionType").enummodule
  UserCommitFilesAction = ::Google::Protobuf::DescriptorPool.generated_pool.lookup("gitaly.UserCommitFilesAction").msgclass
  UserCommitFilesRequestHeader = ::Google::Protobuf::DescriptorPool.generated_pool.lookup("gitaly.UserCommitFilesRequestHeader").msgclass
  UserCommitFilesRequest = ::Google::Protobuf::DescriptorPool.generated_pool.lookup("gitaly.UserCommitFilesRequest").msgclass
  UserCommitFilesResponse = ::Google::Protobuf::DescriptorPool.generated_pool.lookup("gitaly.UserCommitFilesResponse").msgclass
  UserRebaseConfirmableRequest = ::Google::Protobuf::DescriptorPool.generated_pool.lookup("gitaly.UserRebaseConfirmableRequest").msgclass
  UserRebaseConfirmableRequest::Header = ::Google::Protobuf::DescriptorPool.generated_pool.lookup("gitaly.UserRebaseConfirmableRequest.Header").msgclass
  UserRebaseConfirmableResponse = ::Google::Protobuf::DescriptorPool.generated_pool.lookup("gitaly.UserRebaseConfirmableResponse").msgclass
  UserSquashRequest = ::Google::Protobuf::DescriptorPool.generated_pool.lookup("gitaly.UserSquashRequest").msgclass
  UserSquashResponse = ::Google::Protobuf::DescriptorPool.generated_pool.lookup("gitaly.UserSquashResponse").msgclass
  UserApplyPatchRequest = ::Google::Protobuf::DescriptorPool.generated_pool.lookup("gitaly.UserApplyPatchRequest").msgclass
  UserApplyPatchRequest::Header = ::Google::Protobuf::DescriptorPool.generated_pool.lookup("gitaly.UserApplyPatchRequest.Header").msgclass
  UserApplyPatchResponse = ::Google::Protobuf::DescriptorPool.generated_pool.lookup("gitaly.UserApplyPatchResponse").msgclass
  UserUpdateSubmoduleRequest = ::Google::Protobuf::DescriptorPool.generated_pool.lookup("gitaly.UserUpdateSubmoduleRequest").msgclass
  UserUpdateSubmoduleResponse = ::Google::Protobuf::DescriptorPool.generated_pool.lookup("gitaly.UserUpdateSubmoduleResponse").msgclass
end
