// Code generated by sqlc. DO NOT EDIT.

package sqlc

import ()

type MssmtNode struct {
	HashKey  []byte `json:"hash_key"`
	LHashKey []byte `json:"l_hash_key"`
	RHashKey []byte `json:"r_hash_key"`
	Key      []byte `json:"key"`
	Value    []byte `json:"value"`
	Sum      int64  `json:"sum"`
}

type MssmtRoot struct {
	RootHash []byte `json:"root_hash"`
}

type NftDatum struct {
	ID   string `json:"id"`
	Url  string `json:"url"`
	Memo string `json:"memo"`
}
