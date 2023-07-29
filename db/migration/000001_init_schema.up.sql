CREATE TABLE IF NOT EXISTS nft_nodes (
                                           hash_key BYTEA NOT NULL,

                                           l_hash_key BYTEA,

                                           r_hash_key BYTEA,

                                           key BYTEA,

                                           value BYTEA,

                                           PRIMARY KEY (hash_key)
    );

CREATE INDEX IF NOT EXISTS nft_nodes_l_hash_key_idx ON nft_nodes (l_hash_key);
CREATE INDEX IF NOT EXISTS nft_nodes_r_hash_key_idx ON nft_nodes (r_hash_key);

CREATE TABLE IF NOT EXISTS nft_roots (
                                           root_hash BYTEA NOT NULL,

                                           FOREIGN KEY ( root_hash) REFERENCES nft_nodes ( hash_key) ON DELETE CASCADE
    );


CREATE TABLE IF NOT EXISTS nft_data (
    id VARCHAR NOT NULL PRIMARY KEY,
    url VARCHAR NOT NULL,
    memo VARCHAR NOT NULL,
    txId VARCHAR NOT NULL
);