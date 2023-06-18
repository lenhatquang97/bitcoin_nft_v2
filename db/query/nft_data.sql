-- name: InsertNftData :exec
INSERT INTO nft_data (
    id,
    url,
    memo,
    txId
) VALUES (
    $1, $2, $3, $4
         );

-- name: GetListNft :many
SELECT *
FROM nft_data
LIMIT $1 OFFSET $2;

-- name: GetAllNft :many
SELECT *
FROM nft_data;

-- name: GetNftDataByID :one
SELECT *
FROM nft_data
WHERE id=$1
LIMIT 1;

-- name: GetNFtDataByUrl :one
SELECT *
FROM nft_data
WHERE url = $1
LIMIT 1;

-- name: DeleteNftDataByUrl :exec
DELETE FROM nft_data WHERE url=$1;