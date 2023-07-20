package handler

import (
	"bitcoin_nft_v2/business"
	"bitcoin_nft_v2/ipfs"
	"encoding/hex"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

const (
	NFT_LEN_FORMAT = 3
)

type ErrorResponse struct {
	Code    int32  `json:"code"`
	Message string `json:"message"`
}

func WrapperErrorMsgResponse(code int32, msg string) *ErrorResponse {
	return &ErrorResponse{
		Code:    code,
		Message: msg,
	}
}

func WrapperSend(ctx *gin.Context) {
	var req SendRequest
	err := ctx.ShouldBindJSON(&req)
	if err != nil {
		ctx.JSON(500, WrapperErrorMsgResponse(500, err.Error()))
		fmt.Println(err)
		return
	}

	if len(req.Urls) == 0 && req.Address != "default" {
		fmt.Println(err)
		ctx.JSON(400, WrapperErrorMsgResponse(400, "Nft Url must not be empty"))
		return
	}

	if len(req.Passphrase) == 0 {
		ctx.JSON(400, WrapperErrorMsgResponse(400, "Passphrase must not be empty"))
		return
	}

	var data business.NftData

	if sv.Mode == OFF_CHAIN {
		if len(req.Data) != NFT_LEN_FORMAT {
			ctx.JSON(400, WrapperErrorMsgResponse(400, "Data is empty"))
			return
		}

		for _, item := range req.Data {
			if item == "" {
				ctx.JSON(400, WrapperErrorMsgResponse(400, "Data is empty"))
				return
			}
		}

		data.ID = req.Data[0]
		data.Url = req.Data[1]
		data.Memo = req.Data[2]
	}

	// check for mode on chain
	commitTxId, revealTxId, fee, err := sv.Send(req.Address, req.IsSendNFT, req.IsRef, req.Urls, req.TxID, data, req.Passphrase)
	if err != nil {
		fmt.Println(err)
		ctx.JSON(500, WrapperErrorMsgResponse(500, err.Error()))
		return
	}

	ctx.JSON(200, &SendResponse{
		Code:    200,
		Message: "OK",
		Data: SendResponseData{
			CommitTxID: commitTxId,
			RevealTxID: revealTxId,
			Fee:        fee,
		},
	})
}

func WrapperPredefineEstimatedFee(ctx *gin.Context) {
	var req SendRequest
	err := ctx.ShouldBindJSON(&req)
	if err != nil {
		ctx.JSON(500, err)
		fmt.Println(err)
		return
	}

	if len(req.Urls) == 0 {
		fmt.Println(err)
		ctx.JSON(400, "Nft Url must not be empty")
		return
	}

	if len(req.Passphrase) == 0 {
		ctx.JSON(400, "Passphrase must not be empty")
		return
	}

	// check for mode on chain
	var data []string
	if sv.Mode == ON_CHAIN {
		data = req.Urls
	} else {
		data = req.Data
	}
	fee, err := sv.CalculateFee(req.Address, req.IsRef, req.IsMint, data, req.Passphrase)
	if err != nil {
		fmt.Println(err)
		ctx.JSON(500, err)
		return
	}

	ctx.JSON(200, fee)
}

func WrapperImportProof(ctx *gin.Context) {
	var req ImportNftDataRequest
	err := ctx.ShouldBindJSON(&req)
	if err != nil {
		ctx.JSON(500, WrapperErrorMsgResponse(500, err.Error()))
		return
	}

	if req.ID == "" || req.Url == "" || req.Memo == "" {
		ctx.JSON(400, WrapperErrorMsgResponse(400, "Input invalid"))
		return
	}

	err = sv.ImportProof(req.ID, req.Url, req.Memo)
	if err != nil {
		fmt.Println(err)
		ctx.JSON(400, WrapperErrorMsgResponse(400, err.Error()))
		return
	}

	ctx.JSON(200, WrapperErrorMsgResponse(200, "OK"))
}

func WrapperExportProof(ctx *gin.Context) {
	var req ExportNftDataRequest
	err := ctx.ShouldBindJSON(&req)
	if err != nil {
		ctx.JSON(500, WrapperErrorMsgResponse(500, err.Error()))
		return
	}

	if req.Url == "" {
		ctx.JSON(400, WrapperErrorMsgResponse(400, "Input invalid"))
		return
	}

	data, err := sv.ExportProof(req.Url)
	if err != nil {
		ctx.JSON(400, WrapperErrorMsgResponse(400, err.Error()))
		return
	}

	ctx.JSON(200, &ExportProofResponse{
		Code:    200,
		Message: "OK",
		Data: business.NftData{
			ID:   data.ID,
			Url:  data.Url,
			Memo: data.Memo,
		},
	})
}

func WrapperCheckBalance(ctx *gin.Context) {
	var req CheckBalanceRequest
	err := ctx.ShouldBindJSON(&req)
	if err != nil {
		ctx.JSON(500, WrapperErrorMsgResponse(500, err.Error()))
	}
	if req.Address == "" {
		ctx.JSON(400, WrapperErrorMsgResponse(400, "Input invalid"))
		return
	}

	balance, err := sv.CheckBalance(req.Address)
	if err != nil {
		ctx.JSON(400, WrapperErrorMsgResponse(400, err.Error()))
		return
	}

	ctx.JSON(200, &CheckBalanceResponse{
		Code:    200,
		Message: "OK",
		Data:    int64(balance),
	})
}

func WrapperViewNftData(ctx *gin.Context) {
	var req ViewNftDataRequest
	err := ctx.ShouldBindQuery(&req)
	if err != nil {
		ctx.JSON(500, WrapperErrorMsgResponse(500, err.Error()))
		return
	}

	nftData, err := sv.ViewNftData()
	if err != nil {
		ctx.JSON(400, WrapperErrorMsgResponse(400, err.Error()))
		return
	}

	var items []business.NftData
	for _, item := range nftData {
		items = append(items, business.NftData{
			ID:     item.ID,
			Url:    item.Url,
			Memo:   item.Memo,
			TxID:   item.TxID,
			Binary: item.Binary,
		})
	}

	res := &ViewNftDataResponse{
		Code:    200,
		Message: "OK",
		Data:    items,
	}

	ctx.JSON(200, res)
}

func WrapperSetMode(ctx *gin.Context) {
	var req SwitchModeRequest
	err := ctx.ShouldBindJSON(&req)
	if err != nil {
		ctx.JSON(400, err)
		return
	}

	if req.Mode == "" {
		ctx.JSON(400, WrapperErrorMsgResponse(400, "Mode just only is on_chain OR off_chain"))
		return
	}

	err = sv.SetMode(req.Mode)
	if err != nil {
		ctx.JSON(400, WrapperErrorMsgResponse(400, err.Error()))
		return
	}

	ctx.JSON(200, "OK")
}

func WrapperCreateWallet(ctx *gin.Context) {
	var req CreateWalletRequest
	err := ctx.ShouldBindJSON(&req)
	if err != nil {
		ctx.JSON(400, WrapperErrorMsgResponse(400, err.Error()))
		return
	}

	if req.Name == "" || req.Passphrase == "" {
		ctx.JSON(400, WrapperErrorMsgResponse(400, "Input invalid"))
		return
	}

	seed, err := sv.CreateWallet(req.Passphrase)
	if err != nil {
		ctx.JSON(400, WrapperErrorMsgResponse(400, err.Error()))
		return
	}

	ctx.JSON(200, &CreateWalletResponse{
		Code:    200,
		Message: "OK",
		Data: &CreateWalletResponseData{
			Seed: seed,
		},
	})
}

func WrapperGetNftFromUtxo(ctx *gin.Context) {
	res, txIds, originTxIds, err := sv.GetAllNfts()
	if err != nil {
		fmt.Println(err)
		ctx.JSON(400, WrapperErrorMsgResponse(400, err.Error()))
	}

	var outputRes []NftFromUtxo
	for i, item := range res {
		tmpMimeType := http.DetectContentType(item)
		tmpNft := NftFromUtxo{
			HexData:    hex.EncodeToString(item),
			MimeType:   tmpMimeType,
			TxId:       txIds[i],
			OriginTxId: originTxIds[i],
		}
		outputRes = append(outputRes, tmpNft)
	}

	//Hardcoded NFT
	// hardCodedNft := NftFromUtxo{
	// 	HexData:  "31323334350a31323334350a6e6f0a6e6f0a4f4b0a",
	// 	MimeType: "text/plain; charset=utf-8",
	// 	TxId:     "f2b49031b2aecc5c87cd71a3dce67bd89201493610f9e329f920968f86402358",
	// }

	// hardCodedMp3 := NftFromUtxo{
	// 	HexData:  SAMPLE_MP3,
	// 	MimeType: "audio/mpeg",
	// 	TxId:     "f2b49031b2aecc5c87cd71a3djs67bd89201493610f9e329f920968f86402358",
	// }
	// outputRes = append(outputRes, hardCodedNft)
	// outputRes = append(outputRes, hardCodedMp3)

	ctx.JSON(200, &GetNftFromUtxoRes{
		Code:    200,
		Message: "OK",
		Data:    outputRes,
	})
}

func WrapperGetTxSize(ctx *gin.Context) {
	type getTxSizeReq struct {
		TxID string `json:"txId" form:"txId"`
	}

	type getTxSizeRes struct {
		VirtualSize   int64 `json:"virtualSize"`
		SerializeSize int   `json:"serializeSize"`
	}
	var req getTxSizeReq
	err := ctx.ShouldBindQuery(&req)
	if err != nil {
		ctx.JSON(400, WrapperErrorMsgResponse(400, err.Error()))
		return
	}

	if req.TxID == "" {
		ctx.JSON(400, WrapperErrorMsgResponse(400, "TX_ID_REQUIRED"))
		return
	}

	virtualSize, serializeSize, err := sv.GetTxSize(req.TxID)
	if err != nil {
		ctx.JSON(400, WrapperErrorMsgResponse(400, err.Error()))
		return
	}

	ctx.JSON(200, &getTxSizeRes{
		VirtualSize:   virtualSize,
		SerializeSize: serializeSize,
	})
}

func WrapperRenderTree(ctx *gin.Context) {
	err := sv.RenderTree()
	if err != nil {
		ctx.JSON(400, WrapperErrorMsgResponse(400, err.Error()))
		return
	}

	ctx.JSON(200, &struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}{
		Code:    200,
		Message: "OK",
	})
}

func WrapperIpfsLink(ctx *gin.Context) {
	filePath := ctx.Query("filePath")
	if len(filePath) == 0 {
		ctx.JSON(400, WrapperErrorMsgResponse(400, "filePath is required"))
		return
	}
	fileLink, err := ipfs.GetIpfsLink(filePath)
	if err != nil {
		ctx.JSON(400, WrapperErrorMsgResponse(400, err.Error()))
		return
	}

	type getIPFSUrlRes struct {
		Url string `json:"url"`
	}
	ctx.JSON(200, &getIPFSUrlRes{
		Url: fileLink,
	})
}
