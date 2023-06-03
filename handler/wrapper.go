package handler

import "github.com/gin-gonic/gin"

func WrapperSend(ctx *gin.Context) {
	var req SendRequest
	err := ctx.ShouldBindJSON(&req)
	if err != nil {
		ctx.JSON(500, err)
	}

	if len(req.Urls) == 0 {
		ctx.JSON(400, "Nft Url must not be empty")
	}

	err = sv.Send(req.Address, req.Amount, req.Urls)
	if err != nil {
		ctx.JSON(500, err)
	}

	ctx.JSON(200, "OK")
}

func WrapperImportProof(ctx *gin.Context) {
	var req ImportNftDataRequest
	err := ctx.ShouldBindJSON(&req)
	if err != nil {
		ctx.JSON(500, err)
	}

	if req.ID == "" || req.Url == "" || req.Memo == "" {
		ctx.JSON(400, "Input invalid")
	}

	err = sv.ImportProof(req.ID, req.Url, req.Memo)
	if err != nil {
		ctx.JSON(400, err)
	}

	ctx.JSON(200, "OK")
}

func WrapperExportProof(ctx *gin.Context) {
	var req ExportNftDataRequest
	err := ctx.ShouldBindJSON(&req)
	if err != nil {
		ctx.JSON(500, err)
	}

	if req.Url == "" {
		ctx.JSON(400, "Input invalid")
	}

	data, err := sv.ExportProof(req.Url)
	if err != nil {
		ctx.JSON(400, err)
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
	err := ctx.ShouldBindQuery(&req)
	if err != nil {
		ctx.JSON(500, err)
	}

	if req.Address == "" {
		ctx.JSON(400, "Input invalid")
	}

	balance, err := sv.CheckBalance(req.Address)
	if err != nil {
		ctx.JSON(400, err)
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
	}

	nftData, err := sv.ViewNftData()
	if err != nil {
		ctx.JSON(400, err)
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
