package sqlc

import (
	"context"
)

type Querier interface {
	FetchChildren(ctx context.Context, arg FetchChildrenParams) ([]FetchChildrenRow, error)
	FetchChildrenSelfJoin(ctx context.Context, arg FetchChildrenSelfJoinParams) ([]FetchChildrenSelfJoinRow, error)
	FetchRootNode(ctx context.Context, namespace string) (MssmtNode, error)
	InsertBranch(ctx context.Context, arg InsertBranchParams) error
	InsertCompactedLeaf(ctx context.Context, arg InsertCompactedLeafParams) error
	InsertLeaf(ctx context.Context, arg InsertLeafParams) error
	UpsertRootNode(ctx context.Context, arg UpsertRootNodeParams) error
	DeleteNode(ctx context.Context, arg DeleteNodeParams) (int64, error)
}

var _ Querier = (*Queries)(nil)
