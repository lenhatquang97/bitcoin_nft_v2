-- name: InsertNftData :exec
INSERT INTO nft_data (
    id,
    url,
    memo
) VALUES (
    $1, $2, $3
         );

-- name: GetListNft :many
SELECT *
FROM nft_data
LIMIT $1;

-- name: GetNftData :one
SELECT *
FROM nft_data
WHERE id=$1
LIMIT 1;