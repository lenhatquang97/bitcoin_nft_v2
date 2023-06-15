-- name: InsertBranch :exec
INSERT INTO nft_nodes (
    hash_key, l_hash_key, r_hash_key, key, value, sum
) VALUES ($1, $2, $3, NULL, NULL, $4);

-- name: InsertLeaf :exec
INSERT INTO nft_nodes (
    hash_key, l_hash_key, r_hash_key, key, value, sum
) VALUES ($1, NULL, NULL, NULL, $2, $3);

-- name: InsertCompactedLeaf :exec
INSERT INTO nft_nodes (
    hash_key, l_hash_key, r_hash_key, key, value, sum
) VALUES ($1, NULL, NULL, $2, $3, $4);

-- name: FetchChildren :many
WITH RECURSIVE mssmt_branches_cte (
                                   hash_key, l_hash_key, r_hash_key, key, value, sum, depth
    )
                   AS (
        SELECT r.hash_key, r.l_hash_key, r.r_hash_key, r.key, r.value, r.sum, 0 as depth
        FROM nft_nodes r
        WHERE r.hash_key = $1
        UNION ALL
        SELECT n.hash_key, n.l_hash_key, n.r_hash_key, n.key, n.value, n.sum, depth+1
        FROM nft_nodes n, mssmt_branches_cte b
        WHERE n.hash_key=b.l_hash_key OR n.hash_key=b.r_hash_key
    ) SELECT * FROM mssmt_branches_cte WHERE depth < 3;


-- name: FetchChildrenSelfJoin :many
WITH subtree_cte (
                  hash_key, l_hash_key, r_hash_key, key, value, sum, depth
    ) AS (
    SELECT r.hash_key, r.l_hash_key, r.r_hash_key, r.key, r.value, r.sum, 0 as depth
    FROM nft_nodes r
    WHERE r.hash_key = $1
    UNION ALL
    SELECT c.hash_key, c.l_hash_key, c.r_hash_key, c.key, c.value, c.sum, depth+1
    FROM nft_nodes c
             INNER JOIN subtree_cte r ON r.l_hash_key=c.hash_key OR r.r_hash_key=c.hash_key
) SELECT * from subtree_cte WHERE depth < 3;

-- name: DeleteNode :execrows
DELETE FROM nft_nodes WHERE hash_key = $1;

-- name: FetchRootNode :one
SELECT nodes.*
FROM nft_nodes nodes
         JOIN nft_roots roots
              ON roots.root_hash = nodes.hash_key;

-- name: UpsertRootNode :exec
INSERT INTO nft_roots (
    root_hash
) VALUES (
             $1
         );

-- name: GetAllNodeByNameSpace :many
SELECT * from nft_nodes;