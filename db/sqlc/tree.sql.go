// Code generated by sqlc. DO NOT EDIT.
// source: tree.sql

package sqlc

import (
	"context"
)

const deleteNode = `-- name: DeleteNode :execrows
DELETE FROM nft_nodes WHERE hash_key = $1
`

func (q *Queries) DeleteNode(ctx context.Context, hashKey []byte) (int64, error) {
	result, err := q.db.ExecContext(ctx, deleteNode, hashKey)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

const fetchChildren = `-- name: FetchChildren :many
WITH RECURSIVE mssmt_branches_cte (
                                   hash_key, l_hash_key, r_hash_key, key, value, depth
    )
                   AS (
        SELECT r.hash_key, r.l_hash_key, r.r_hash_key, r.key, r.value, 0 as depth
        FROM nft_nodes r
        WHERE r.hash_key = $1
        UNION ALL
        SELECT n.hash_key, n.l_hash_key, n.r_hash_key, n.key, n.value, depth+1
        FROM nft_nodes n, mssmt_branches_cte b
        WHERE n.hash_key=b.l_hash_key OR n.hash_key=b.r_hash_key
    ) SELECT hash_key, l_hash_key, r_hash_key, key, value, depth FROM mssmt_branches_cte WHERE depth < 3
`

type FetchChildrenRow struct {
	HashKey  []byte      `json:"hash_key"`
	LHashKey []byte      `json:"l_hash_key"`
	RHashKey []byte      `json:"r_hash_key"`
	Key      []byte      `json:"key"`
	Value    []byte      `json:"value"`
	Depth    interface{} `json:"depth"`
}

func (q *Queries) FetchChildren(ctx context.Context, hashKey []byte) ([]FetchChildrenRow, error) {
	rows, err := q.db.QueryContext(ctx, fetchChildren, hashKey)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []FetchChildrenRow
	for rows.Next() {
		var i FetchChildrenRow
		if err := rows.Scan(
			&i.HashKey,
			&i.LHashKey,
			&i.RHashKey,
			&i.Key,
			&i.Value,
			&i.Depth,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const fetchChildrenSelfJoin = `-- name: FetchChildrenSelfJoin :many
WITH subtree_cte (
                  hash_key, l_hash_key, r_hash_key, key, value, depth
    ) AS (
    SELECT r.hash_key, r.l_hash_key, r.r_hash_key, r.key, r.value, 0 as depth
    FROM nft_nodes r
    WHERE r.hash_key = $1
    UNION ALL
    SELECT c.hash_key, c.l_hash_key, c.r_hash_key, c.key, c.value, depth+1
    FROM nft_nodes c
             INNER JOIN subtree_cte r ON r.l_hash_key=c.hash_key OR r.r_hash_key=c.hash_key
) SELECT hash_key, l_hash_key, r_hash_key, key, value, depth from subtree_cte WHERE depth < 3
`

type FetchChildrenSelfJoinRow struct {
	HashKey  []byte      `json:"hash_key"`
	LHashKey []byte      `json:"l_hash_key"`
	RHashKey []byte      `json:"r_hash_key"`
	Key      []byte      `json:"key"`
	Value    []byte      `json:"value"`
	Depth    interface{} `json:"depth"`
}

func (q *Queries) FetchChildrenSelfJoin(ctx context.Context, hashKey []byte) ([]FetchChildrenSelfJoinRow, error) {
	rows, err := q.db.QueryContext(ctx, fetchChildrenSelfJoin, hashKey)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []FetchChildrenSelfJoinRow
	for rows.Next() {
		var i FetchChildrenSelfJoinRow
		if err := rows.Scan(
			&i.HashKey,
			&i.LHashKey,
			&i.RHashKey,
			&i.Key,
			&i.Value,
			&i.Depth,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const fetchRootNode = `-- name: FetchRootNode :one
SELECT nodes.hash_key, nodes.l_hash_key, nodes.r_hash_key, nodes.key, nodes.value
FROM nft_nodes nodes
         JOIN nft_roots roots
              ON roots.root_hash = nodes.hash_key
`

func (q *Queries) FetchRootNode(ctx context.Context) (NftNode, error) {
	row := q.db.QueryRowContext(ctx, fetchRootNode)
	var i NftNode
	err := row.Scan(
		&i.HashKey,
		&i.LHashKey,
		&i.RHashKey,
		&i.Key,
		&i.Value,
	)
	return i, err
}

const getAllNodeByNameSpace = `-- name: GetAllNodeByNameSpace :many
SELECT hash_key, l_hash_key, r_hash_key, key, value from nft_nodes
`

func (q *Queries) GetAllNodeByNameSpace(ctx context.Context) ([]NftNode, error) {
	rows, err := q.db.QueryContext(ctx, getAllNodeByNameSpace)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []NftNode
	for rows.Next() {
		var i NftNode
		if err := rows.Scan(
			&i.HashKey,
			&i.LHashKey,
			&i.RHashKey,
			&i.Key,
			&i.Value,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const insertBranch = `-- name: InsertBranch :exec
INSERT INTO nft_nodes (
    hash_key, l_hash_key, r_hash_key, key, value
) VALUES ($1, $2, $3, NULL, NULL)
`

type InsertBranchParams struct {
	HashKey  []byte `json:"hash_key"`
	LHashKey []byte `json:"l_hash_key"`
	RHashKey []byte `json:"r_hash_key"`
}

func (q *Queries) InsertBranch(ctx context.Context, arg InsertBranchParams) error {
	_, err := q.db.ExecContext(ctx, insertBranch, arg.HashKey, arg.LHashKey, arg.RHashKey)
	return err
}

const insertCompactedLeaf = `-- name: InsertCompactedLeaf :exec
INSERT INTO nft_nodes (
    hash_key, l_hash_key, r_hash_key, key, value
) VALUES ($1, NULL, NULL, $2, $3)
`

type InsertCompactedLeafParams struct {
	HashKey []byte `json:"hash_key"`
	Key     []byte `json:"key"`
	Value   []byte `json:"value"`
}

func (q *Queries) InsertCompactedLeaf(ctx context.Context, arg InsertCompactedLeafParams) error {
	_, err := q.db.ExecContext(ctx, insertCompactedLeaf, arg.HashKey, arg.Key, arg.Value)
	return err
}

const insertLeaf = `-- name: InsertLeaf :exec
INSERT INTO nft_nodes (
    hash_key, l_hash_key, r_hash_key, key, value
) VALUES ($1, NULL, NULL, NULL, $2)
`

type InsertLeafParams struct {
	HashKey []byte `json:"hash_key"`
	Value   []byte `json:"value"`
}

func (q *Queries) InsertLeaf(ctx context.Context, arg InsertLeafParams) error {
	_, err := q.db.ExecContext(ctx, insertLeaf, arg.HashKey, arg.Value)
	return err
}

const upsertRootNode = `-- name: UpsertRootNode :exec
INSERT INTO nft_roots (
    root_hash
) VALUES (
             $1
         )
`

func (q *Queries) UpsertRootNode(ctx context.Context, rootHash []byte) error {
	_, err := q.db.ExecContext(ctx, upsertRootNode, rootHash)
	return err
}
