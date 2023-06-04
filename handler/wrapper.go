package handler

import (
	"fmt"
	"github.com/gin-gonic/gin"
)

func WrapperSend(ctx *gin.Context) {
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
	err = sv.Send(req.Address, req.Amount, req.Urls, req.Passphrase)
	if err != nil {
		fmt.Println(err)
		ctx.JSON(500, err)
		return
	}

	ctx.JSON(200, "OK")
}

func WrapperImportProof(ctx *gin.Context) {
	var req ImportNftDataRequest
	err := ctx.ShouldBindJSON(&req)
	if err != nil {
		ctx.JSON(500, err)
		return
	}

	if req.ID == "" || req.Url == "" || req.Memo == "" {
		ctx.JSON(400, "Input invalid")
		return
	}

	err = sv.ImportProof(req.ID, req.Url, req.Memo)
	if err != nil {
		fmt.Println(err)
		ctx.JSON(400, err)
		return
	}

	ctx.JSON(200, "OK")
}

func WrapperExportProof(ctx *gin.Context) {
	var req ExportNftDataRequest
	err := ctx.ShouldBindJSON(&req)
	if err != nil {
		ctx.JSON(500, err)
		return
	}

	if req.Url == "" {
		ctx.JSON(400, "Input invalid")
		return
	}

	data, err := sv.ExportProof(req.Url)
	if err != nil {
		ctx.JSON(400, err)
		return
	}

	ctx.JSON(200, &ExportProofResponse{
		Code:    "200",
		Message: "OK",
		Data: NftData{
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
		ctx.JSON(500, err)
	}
	if req.Address == "" {
		ctx.JSON(400, "Input invalid")
		return
	}

	balance, err := sv.CheckBalance(req.Address)
	if err != nil {
		ctx.JSON(400, err)
		return
	}

	ctx.JSON(200, &CheckBalanceResponse{
		Code:    "200",
		Message: "OK",
		Data:    int64(balance),
	})
}

func WrapperViewNftData(ctx *gin.Context) {
	var req ViewNftDataRequest
	err := ctx.ShouldBindJSON(&req)
	if err != nil {
		ctx.JSON(500, err)
		return
	}

	nftData, err := sv.ViewNftData()
	if err != nil {
		ctx.JSON(400, err)
		return
	}

	var items []NftData
	for _, item := range nftData {
		items = append(items, NftData{
			ID:   item.ID,
			Url:  item.Url,
			Memo: item.Memo,
		})
	}

	res := &ViewNftDataResponse{
		Code:    "200",
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
		ctx.JSON(400, "Mode just only is on_chain OR off_chain")
		return
	}

	err = sv.SetMode(req.Mode)
	if err != nil {
		ctx.JSON(400, err)
		return
	}

	ctx.JSON(200, "OK")
}

func WrapperCreateWallet(ctx *gin.Context) {
	var req CreateWalletRequest
	err := ctx.ShouldBindJSON(&req)
	if err != nil {
		ctx.JSON(400, err)
		return
	}

	if req.Name == "" || req.Passphrase == "" {
		ctx.JSON(400, "Input invalid")
		return
	}

	err = sv.CreateWallet(req.Name, req.Passphrase)
	if err != nil {
		ctx.JSON(400, err)
		return
	}

	ctx.JSON(200, "OK")
}

func WrapperGetTx(ctx *gin.Context) {

	var req GetTxRequest
	err := ctx.ShouldBindQuery(&req)
	if err != nil {
		fmt.Println(err)
	}

	if req.TxID == "" {
		fmt.Println(ctx.Params)
		ctx.JSON(400, "txId is required")
		return
	}

	data, err := sv.GetTx(req.TxID)
	if err != nil {
		ctx.JSON(400, err.Error())
	}

	ctx.JSON(200, data)
}
