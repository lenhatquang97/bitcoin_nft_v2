CREATE TABLE IF NOT EXISTS mssmt_nodes (
    -- hash_key is the hash key by which we reference all nodes.
                                           hash_key BYTEA NOT NULL,

    -- l_hash_key is the hash key of the left child or NULL. If this is a
    -- branch then either l_hash_key or r_hash_key is not NULL.
                                           l_hash_key BYTEA,

    -- r_hash_key is the hash key of the right child or NULL. If this is a
    -- branch then either l_hash_key or r_hash_key is not NULL.
                                           r_hash_key BYTEA,

    -- key is the leaf key if this is a compacted leaf node.
                                           key BYTEA,

    -- value is the leaf value if this is a leaf node.
                                           value BYTEA,

    -- sum is the sum of the node.
                                           sum BIGINT NOT NULL,

    -- namespace allows an application to store several distinct MS-SMT nodes
    -- in the same table, partitioning them by the namespace value.
--                                            namespace VARCHAR NOT NULL,

    -- A combination of the hash_key and the namespace comprise our primary
    -- key. Using these two in concert allows us to do things like copy trees
    -- between namespaces.
--                                            PRIMARY KEY (hash_key, namespace)
                                           PRIMARY KEY (hash_key)
    );

CREATE INDEX IF NOT EXISTS mssmt_nodes_l_hash_key_idx ON mssmt_nodes (l_hash_key);
CREATE INDEX IF NOT EXISTS mssmt_nodes_r_hash_key_idx ON mssmt_nodes (r_hash_key);

CREATE TABLE IF NOT EXISTS mssmt_roots (
    -- namespace allows us to store several root hash pointers for distinct
    -- trees.
--                                            namespace VARCHAR NOT NULL PRIMARY KEY,

    -- root_hash points to the root hash node of the MS-SMT tree.
                                           root_hash BYTEA NOT NULL,

--                                            FOREIGN KEY (namespace, root_hash) REFERENCES mssmt_nodes (namespace, hash_key) ON DELETE CASCADE
                                           FOREIGN KEY ( root_hash) REFERENCES mssmt_nodes ( hash_key) ON DELETE CASCADE
    );


CREATE TABLE IF NOT EXISTS nft_data (
    id VARCHAR NOT NULL PRIMARY KEY,
    url VARCHAR NOT NULL,
    memo VARCHAR NOT NULL
);