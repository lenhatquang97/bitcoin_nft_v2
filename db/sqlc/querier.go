package sqlc

import (
	"context"
)

type Querier interface {
	FetchChildren(ctx context.Context, hashKey []byte) ([]FetchChildrenRow, error)
	FetchChildrenSelfJoin(ctx context.Context, hashKey []byte) ([]FetchChildrenSelfJoinRow, error)
	FetchRootNode(ctx context.Context) (MssmtNode, error)
	InsertBranch(ctx context.Context, arg InsertBranchParams) error
	InsertCompactedLeaf(ctx context.Context, arg InsertCompactedLeafParams) error
	InsertLeaf(ctx context.Context, arg InsertLeafParams) error
	UpsertRootNode(ctx context.Context, hashKey []byte) error
	DeleteNode(ctx context.Context, hashKey []byte) (int64, error)
}

var _ Querier = (*Queries)(nil)
