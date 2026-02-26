package console

import (
	"crypto/x509"
	"time"

	"gitee.com/we7coreteam/k8s-offline/common/helper"
	"gitee.com/we7coreteam/k8s-offline/common/service/config"
)

type PayInfo struct {
	Ticket string `json:"ticket"`
}
type PayTicketInfo struct {
	PayInfo PayInfo `json:"payinfo"` //程序 返回error字段
}
type CouponCode struct {
	Code string `json:"code"`
}

/*
*

	'canBuy' => $canBuy,
	                   'needCheckFinish' => true,
	                   'needCheckAfter' => false,
	                   'orderSn' => $k3kOrder->ip_order_sn,
	                   'orderId' => $k3kOrder->ip_order_id,
	                   "error" => "有未验收的订单，请先完成验收后再购买资源包",
	                   "goBtn" => "去验收"
*/
type LastPaidOrder struct {
	CanBuy          bool     `json: "canBuy"`
	NeedCheckFinish bool     `json: "needCheckFinish"`
	NeedCheckAfter  bool     `json: "needCheckAfter"`
	K3kOrder        K3kOrder `json: "k3kOrder"`
	Error           string   `json: "error"`
	GoBtn           string   `json: "goBtn"`
}

type K3kOrder struct {
	OrderId     int64  `json: "orderId"`
	OrderSn     string `json: "orderSn"`
	OrderStatus string `json: "orderStatus"`
	ReturnAt    string `json: "returnAt"` //退业务时间
	BuyMode     string `json: "buymode"`  //base 基础购买 expand扩容
	Cpu         int64  `json: "cpu"`      //cpu
	Memory      int64  `json: "memory"`   //内存
	Storage     int64  `json: "storage"`  //存储
	Bandwidth   int64  `json: "bandwidth"`
	Hour        string `json: "hour"` // 购买时长
}

type LastReturnOrder struct {
	HasOrder bool      `json: "hasOrder"`
	K3kOrder *K3kOrder `json: "k3kOrder"`
}

func (c *K3kOrder) GetHour() int64 {
	return helper.FloatStringToInt64(c.Hour)
}

type Coupon struct {
	Code         string `json:"code"`
	Discount     int64  `json:"discount"`
	Groupname    string `json:"groupname"`
	ExpiredAt    string `json:"expiredAt"`
	CanUse       bool   `json:"canuse"`
	Cpu          int64  `json:"cpu"`       // cpu
	Memory       int64  `json:"memory"`    //
	Storage      int64  `json:"storage"`   //
	Bandwidth    int64  `json:"bandwidth"` // 带宽
	Status       string `json:"status"`
	TimeUnit     string `json:"timeunit"`
	TimeQuantity int64  `json:"timequantity"`
}

type ConsoleError struct {
	ErrorMsg string `json:"error"`   //程序 返回error字段
	Message  string `json:"message"` //laravel 返回message字段
}

type LicenseOrder struct {
	Cert string `json:"tls_crt"`
}

func (e *ConsoleError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	return e.ErrorMsg
}

type License struct {
	AppId         string            `json:"appId"`
	AppSecret     string            `json:"appSecret"`
	License       *x509.Certificate `json:"license"`
	FounderSaName string            `json:"founderSaName"` //创始人账号
}

// 当前激活的license信息
var currentLicense *License

func SetCurrentLicense(license *License) {
	currentLicense = license
}
func GetCurrentLicense() *License {
	return currentLicense
}

func IsFree() bool {
	if currentLicense != nil {
		return currentLicense.IsFree()
	}
	return true
}

func RefreshLicense() {
	if licenseClient != nil {
		currentLicense, _ = licenseClient.GetLicense()
		if currentLicense != nil {
			config.MainW7Config, _ = licenseClient.GetConfig(currentLicense.FounderSaName)
		}
	}
}

func init() {
	RefreshLicense()
}

func (l *License) GetAppId() string {
	return l.AppId
}
func (l *License) GetAppSecret() string {
	return l.AppSecret
}
func (l *License) IsFree() bool {
	return "free" == l.GetLicenseType()
}
func (c *License) GetLicenseType() string {
	license := c.License
	licenseType := "free"
	isExpired := false
	if license != nil && len(license.Subject.Province) > 0 {
		licenseType = license.Subject.Province[0]
	}
	if license != nil {
		// endTime = license.NotAfter
		if license.NotAfter.Before(time.Now()) {
			isExpired = true
		}
		if isExpired {
			licenseType = "free"
		}
	}
	return licenseType
}

func (c *License) ToArray() map[string]interface{} {
	license := c.License
	licenseType := "free"
	licenseId := "0"
	if license != nil && len(license.Subject.Province) > 0 {
		licenseType = license.Subject.Province[0]
	}
	endTime := time.Now()
	isExpired := false
	if license != nil {
		licenseId = license.SerialNumber.String()
		endTime = license.NotAfter
		if license.NotAfter.Before(time.Now()) {
			isExpired = true
		}
		if isExpired {
			licenseType = "free"
		}
	}
	return map[string]interface{}{
		"license_type":       licenseType,
		"license_id":         licenseId,
		"license_end_time":   endTime.Format("2006-01-02 15:04:05"),
		"license_is_expired": isExpired,
	}
}

type ThirdPartyCDToken struct {
	Token string `json:"token"`
	// Exp   int64  `json:"expiretime"`
}
