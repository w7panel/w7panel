package controller

import (
	"fmt"
	"log/slog"

	"gitee.com/we7coreteam/k8s-offline/common/service/config"
	"gitee.com/we7coreteam/k8s-offline/common/service/console"
	"gitee.com/we7coreteam/k8s-offline/common/service/k8s"
	"gitee.com/we7coreteam/k8s-offline/common/service/k8s/k3k"
	"gitee.com/we7coreteam/k8s-offline/common/service/k8s/k3k/order"
	"gitee.com/we7coreteam/k8s-offline/common/service/k8s/k3k/types"
	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
	"github.com/we7coreteam/w7-rangine-go/v2/src/http/controller"
)

type Order struct {
	controller.Abstract
}

// 授权购买
func (self Order) CreateLicenseOrder(http *gin.Context) {
	type ParamsValidate struct {
		ProductId string `form:"productId" validate:"required"`
	}
	params := ParamsValidate{}
	if !self.Validate(http, &params) {
		return
	}
	sdkClient, err := console.NewDefaultSdkClient()
	if err != nil {
		self.JsonResponseWithServerError(http, err)
		return
	}
	result, err := sdkClient.CreateDefaultProductOrder(params.ProductId)
	if err != nil {
		self.JsonResponseWithServerError(http, err)
		return
	}
	self.JsonResponseWithoutError(http, result)
	return
}

/*
*
创建资源购买订单
*/
func (self Order) CreateBaseResourceOrder(http *gin.Context) {

	// 初始化参数结构体

	if config.MainW7Config == nil {
		self.JsonResponseWithServerError(http, fmt.Errorf("集群不允许购买资源"))
		return
	}
	// params := types.BuyBaseResource{BaseConfigName: config.MainW7Config.Name}
	params := types.BuyBaseResource{}
	if !self.Validate(http, &params) {
		return
	}

	params.BaseConfigName = config.MainW7Config.Name
	token := http.MustGet("k8s_token").(string)
	k3kUser, err := k3k.TokenToK3kUser(token)
	if err != nil {
		slog.Error("token解析失败", "error", err)
		self.JsonResponseWithServerError(http, err)
		return
	}
	if !k3kUser.IsClusterUser() {
		self.JsonResponseWithServerError(http, fmt.Errorf("集群用户不允许购买资源"))
		return
	}
	payResult, err := order.CreateBaseResourceOrder(&params, k3kUser)
	if err != nil {
		slog.Error("购买失败", "error", err)
		self.JsonResponseWithServerError(http, err)
		return
	}
	self.JsonResponseWithoutError(http, payResult)
	return
}

func (self Order) CreateRenewOrder(http *gin.Context) {

	// 初始化参数结构体

	if config.MainW7Config == nil {
		self.JsonResponseWithServerError(http, fmt.Errorf("集群不允许购买资源"))
		return
	}
	params := types.BuyRenewResource{}
	if !self.Validate(http, &params) {
		return
	}
	params.BaseConfigName = config.MainW7Config.Name
	token := http.MustGet("k8s_token").(string)
	k3kUser, err := k3k.TokenToK3kUser(token)
	if err != nil {
		slog.Error("token解析失败", "error", err)
		self.JsonResponseWithServerError(http, err)
		return
	}
	if !k3kUser.IsClusterUser() {
		self.JsonResponseWithServerError(http, fmt.Errorf("集群用户不允许购买资源"))
		return
	}
	payResult, err := order.CreateRenewOrder(&params, k3kUser)
	if err != nil {
		slog.Error("购买失败", "error", err)
		self.JsonResponseWithServerError(http, err)
		return
	}
	self.JsonResponseWithoutError(http, payResult)
	return
}

func (self Order) CreateExpandOrder(http *gin.Context) {

	// 初始化参数结构体

	if config.MainW7Config == nil {
		self.JsonResponseWithServerError(http, fmt.Errorf("集群不允许购买资源"))
		return
	}
	params := types.BuyExpandResource{}
	if !self.Validate(http, &params) {
		return
	}
	params.BaseConfigName = config.MainW7Config.Name
	token := http.MustGet("k8s_token").(string)
	k3kUser, err := k3k.TokenToK3kUser(token)
	if err != nil {
		slog.Error("token解析失败", "error", err)
		self.JsonResponseWithServerError(http, err)
		return
	}
	if !k3kUser.IsClusterUser() {
		self.JsonResponseWithServerError(http, fmt.Errorf("集群用户不允许购买资源"))
		return
	}
	payResult, err := order.CreateExpandOrder(&params, k3kUser)
	if err != nil {
		slog.Error("购买失败", "error", err)
		self.JsonResponseWithServerError(http, err)
		return
	}
	self.JsonResponseWithoutError(http, payResult)
	return
}

func (self Order) OrderNotify(http *gin.Context) {

	// 初始化参数结构体
	params := types.BuyNotify{}
	if !self.Validate(http, &params) {
		return
	}
	sdk := k8s.NewK8sClient().Sdk
	k3kName := params.K3kName

	sa, err := sdk.GetServiceAccount(sdk.GetNamespace(), k3kName)
	if err != nil {
		slog.Error("获取服务账号失败", "error", err)
		self.JsonResponseWithServerError(http, err)
		return
	}
	k3kUser := types.NewK3kUser(sa)

	orderApi, err := order.NewK3kOrderApi(sdk)
	if err != nil {
		slog.Error("获取订单API失败", "error", err)
		self.JsonResponseWithServerError(http, err)
		return
	}
	err = orderApi.NotifyOrder(k3kUser, params.OrderSn)
	if err != nil {
		slog.Error("通知失败", "error", err)
		self.JsonResponseWithServerError(http, err)
		return
	}
	self.JsonSuccessResponse(http)
	return
}

func (self Order) Refresh(http *gin.Context) {

	// 初始化参数结构体

	token := http.MustGet("k8s_token").(string)
	k3kUser, err := k3k.TokenToK3kUser(token)
	if err != nil {
		slog.Error("token解析失败", "error", err)
		self.JsonResponseWithServerError(http, err)
		return
	}
	if !k3kUser.IsClusterUser() {
		self.JsonResponseWithServerError(http, fmt.Errorf("集群用户不允许购买资源"))
		return
	}
	err = order.Refresh(k3kUser)
	if err != nil {
		slog.Error("购买失败", "error", err)
		self.JsonResponseWithServerError(http, err)
		return
	}
	self.JsonSuccessResponse(http)
	return
}

func (self Order) GetPrice(http *gin.Context) {

	type Result struct {
		Price       decimal.Decimal `json:"price"`
		OriginPrice decimal.Decimal `json:"originPrice"`
		Discount    int64           `json:"discount"`
		BuyMode     string          `json:"buyMode"`
	}
	// 初始化参数结构体
	params := &types.PriceResource{}
	if !self.Validate(http, params) {
		return
	}
	token := http.MustGet("k8s_token").(string)
	k3kUser, err := k3k.TokenToK3kUser(token)
	if err != nil {
		slog.Error("token解析失败", "error", err)
		self.JsonResponseWithServerError(http, err)
		return
	}
	if !k3kUser.IsClusterUser() {
		self.JsonResponseWithServerError(http, fmt.Errorf("集群用户不允许购买资源"))
		return
	}
	compute, err := k3kUser.GetOrderCompute()
	if err != nil {
		slog.Error("获取订单计算失败", "error", err)
		self.JsonResponseWithServerError(http, err)
		return
	}
	computeBase := compute.WithResource(params.BuyResource).WithQuantity(params.UnitQuantity)
	result2 := make(map[string]Result)
	base := Result{
		Price:       computeBase.GetDiscountPrice(order.BASE_BUY),
		OriginPrice: computeBase.GetOriginPrice(),
		Discount:    computeBase.GetDiscount(order.BASE_BUY),
		BuyMode:     order.BASE_BUY,
	}

	renew := Result{
		Price:       compute.GetDiscountPrice(order.RENEW_BUY),
		OriginPrice: compute.GetOriginPrice(),
		Discount:    compute.GetDiscount(order.RENEW_BUY),
		BuyMode:     order.RENEW_BUY,
	}
	expand := Result{BuyMode: order.EXPAND_BUY}
	endTime, err := k3kUser.GetExpireTime()
	if err == nil {
		computeExpand := compute.SubResource(params.BuyResource)
		originPrice, err := computeExpand.GetExpandPrice(endTime)
		if err == nil {
			expand.Price = originPrice
			expand.OriginPrice = originPrice
			expand.Discount = 1.0
		}
	}
	result2[order.BASE_BUY] = base
	result2[order.RENEW_BUY] = renew
	result2[order.EXPAND_BUY] = expand

	self.JsonResponseWithoutError(http, result2)

	return
}

func (self Order) GetConfig(http *gin.Context) {

}
