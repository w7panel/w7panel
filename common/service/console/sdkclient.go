package console

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"math/rand"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"sort"
	"strings"
	"time"

	"gitee.com/we7coreteam/k8s-offline/common/helper"
	resty "github.com/go-resty/resty/v2"

	w7 "github.com/w7corp/sdk-open-cloud-go"
)

type SdkClient struct {
	*w7.Client
	License *License
}

func NewDefaultSdkClient() (*SdkClient, error) {
	licenseClient, err := NewDefaultLicenseClient()
	if err != nil {
		return nil, err
	}
	license, err := licenseClient.GetLicense()
	if err != nil {
		return nil, fmt.Errorf("获取license失败: %w", err)
	}
	return NewSdkClient(license)
}

func NewSdkClient(license *License) (*SdkClient, error) {
	apiUrl := "https://console.w7.cc"
	if helper.IsLocalMock() {
		apiUrl = "http://172.16.1.150:9004"
	}
	client := w7.NewClient(
		license.AppId,
		license.AppSecret,
		w7.Option{
			// ApiUrl: "https://console.w7.cc",
			ApiUrl: apiUrl,
			Debug:  true,
		})
	return &SdkClient{
		Client:  client,
		License: license,
	}, nil
}
func (c *SdkClient) getReq() *resty.Request {
	req := c.Client.GetHttpClient().R().SetHeader("accept", "application/json")
	useragent, ok := os.LookupEnv("USER_AGENT")
	if ok {
		req.SetHeader("User-Agent", useragent).EnableTrace()
	}
	return req
}

// addSignatureToRequest 为 resty 请求添加签名
func (c *SdkClient) addSignatureToRequest(request *resty.Request) error {
	// 从创建 client 时保存的 license 中获取 appId 和 appSecret
	// 这里假设我们可以从某个地方获取到这些信息
	var appId, appSecret string

	// 尝试从 Client 中获取，由于字段是私有的，我们需要其他方式
	// 可以通过反射或者从配置中获取
	// 这里我们使用一个简化的方法，假设可以从某个地方获取到
	// 在实际使用中，你可能需要调整这部分逻辑

	// 由于无法直接访问私有字段，我们需要创建一个新的方法来获取这些信息
	// 或者修改 SdkClient 结构体来保存这些信息
	if c.License != nil {
		appId = c.License.AppId
		appSecret = c.License.AppSecret
	}

	if appId == "" || appSecret == "" {
		return errors.New("appId or appSecret is empty")
	}

	// 使用反射或者重新实现签名逻辑
	// 这里我们重新实现签名逻辑，参考 makeSign 方法
	var sign [16]byte
	if request.Body != nil && resty.DetectContentType(request.Body) == "application/json" {
		body, ok := request.Body.(map[string]interface{})
		if !ok {
			return errors.New("request property body must be map[string]interface{}")
		}
		body["appid"] = appId
		body["timestamp"] = fmt.Sprintf("%d", time.Now().Unix())
		body["nonce"] = c.getRandomString(16)

		signByte, err := json.Marshal(body)
		if err != nil {
			return err
		}
		signStr := string(signByte)
		signStr += appSecret

		hash := md5.Sum([]byte(signStr))
		sign = hash

		body["sign"] = hex.EncodeToString(sign[:])
		request.SetBody(body)
	} else {
		// 处理表单数据
		formData := make(map[string]string)
		if request.FormData != nil {
			for k, v := range request.FormData {
				if len(v) > 0 {
					formData[k] = v[0]
				}
			}
		}

		formData["appid"] = appId
		formData["timestamp"] = fmt.Sprintf("%d", time.Now().Unix())
		formData["nonce"] = c.getRandomString(16)

		var keys []string
		signStr := ""
		for s := range formData {
			if s == "sign" {
				continue
			}
			keys = append(keys, s)
		}
		sort.Strings(keys)
		for i, k := range keys {
			signStr += fmt.Sprintf("%s=%s", k, url.QueryEscape(formData[k]))
			if i < len(keys)-1 {
				signStr += "&"
			}
		}

		signStr += appSecret

		hash := md5.Sum([]byte(signStr))
		sign = hash

		formData["sign"] = hex.EncodeToString(sign[:])
		request.SetFormData(formData)
	}

	return nil
}

// addSignatureToHttpRequest 为 http.Request 添加签名
func (c *SdkClient) addSignatureToHttpRequest(req *http.Request) error {
	// 从创建 client 时保存的 license 中获取 appId 和 appSecret
	var appId, appSecret string
	if c.License != nil {
		appId = c.License.AppId
		appSecret = c.License.AppSecret
	}

	if appId == "" || appSecret == "" {
		return errors.New("appId or appSecret is empty")
	}

	var sign [16]byte

	// 判断请求内容类型
	contentType := req.Header.Get("Content-Type")
	isJSON := strings.Contains(contentType, "application/json")

	if req.Method == "POST" || req.Method == "PUT" {
		// if true {
		// 读取请求体
		body, err := io.ReadAll(req.Body)
		if err != nil {
			return err
		}
		// 重置请求体
		req.Body = io.NopCloser(bytes.NewBuffer(body))

		if isJSON && len(body) > 0 {
			// 处理 JSON 数据
			var jsonData map[string]interface{}
			if err := json.Unmarshal(body, &jsonData); err == nil {
				// 添加签名参数
				jsonData["appid"] = appId
				jsonData["timestamp"] = fmt.Sprintf("%d", time.Now().Unix())
				jsonData["nonce"] = c.getRandomString(16)

				signByte, err := json.Marshal(jsonData)
				if err != nil {
					return err
				}
				signStr := string(signByte)
				signStr += appSecret

				hash := md5.Sum([]byte(signStr))
				sign = hash

				jsonData["sign"] = hex.EncodeToString(sign[:])

				// 重新编码 JSON 并设置到请求体
				jsonBody, err := json.Marshal(jsonData)
				if err != nil {
					return err
				}
				req.Body = io.NopCloser(bytes.NewBuffer(jsonBody))
				req.ContentLength = int64(len(jsonBody))
			}
		} else {
			// 处理表单数据
			formData, err := url.ParseQuery(string(body))
			if err == nil {
				// 添加签名参数到表单数据
				formData.Set("appid", appId)
				formData.Set("timestamp", fmt.Sprintf("%d", time.Now().Unix()))
				formData.Set("nonce", c.getRandomString(16))

				// 生成签名
				var keys []string
				signStr := ""
				for s := range formData {
					if s == "sign" {
						continue
					}
					keys = append(keys, s)
				}
				sort.Strings(keys)
				for i, k := range keys {
					signStr += fmt.Sprintf("%s=%s", k, url.QueryEscape(formData.Get(k)))
					if i < len(keys)-1 {
						signStr += "&"
					}
				}

				signStr += appSecret
				hash := md5.Sum([]byte(signStr))
				sign = hash

				formData.Set("sign", hex.EncodeToString(sign[:]))

				// 重新编码表单数据并设置到请求体
				formDataStr := formData.Encode()
				req.Body = io.NopCloser(strings.NewReader(formDataStr))
				req.ContentLength = int64(len(formDataStr))
				req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			}
		}
	} else {
		// GET 请求，将签名参数添加到查询字符串
		query := req.URL.Query()
		query.Set("appid", appId)
		query.Set("timestamp", fmt.Sprintf("%d", time.Now().Unix()))
		query.Set("nonce", c.getRandomString(16))

		// 生成签名
		var keys []string
		signStr := ""
		for s := range query {
			if s == "sign" {
				continue
			}
			keys = append(keys, s)
		}
		sort.Strings(keys)
		for i, k := range keys {
			signStr += fmt.Sprintf("%s=%s", k, url.QueryEscape(query.Get(k)))
			if i < len(keys)-1 {
				signStr += "&"
			}
		}

		signStr += appSecret
		hash := md5.Sum([]byte(signStr))
		sign = hash

		query.Set("sign", hex.EncodeToString(sign[:]))
		req.URL.RawQuery = query.Encode()
	}

	return nil
}

// getRandomString 生成随机字符串
func (c *SdkClient) getRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

// 支付应用订单创建接口
func (c *SdkClient) CreateDefaultProductOrder(productId string) (*PayTicketInfo, error) {
	result := &PayTicketInfo{}
	errResult := &ConsoleError{}
	params := map[string]string{
		"product_id": productId,
	}
	req := c.getReq()
	res, err := req.SetFormData(params).SetResult(result).SetError(errResult).Post("/api/thirdparty-pay/sdk-license/pay-goods-ip/order/create")
	if err != nil {
		return nil, err
	}
	if res.StatusCode() > 399 {
		// slog.Error("")
		slog.Error("CreateProductOrder", "error", errResult)
		return nil, errResult
	}
	return result, nil
}

func (c *SdkClient) PublishPanelResource(urlValues map[string]string) error {
	req := c.getReq()
	response, err := req.SetFormData(urlValues).Post("/api/thirdparty-cd/k8s-offline/sdk/panel/resource")
	if err != nil {
		return err
	}
	if response.StatusCode() > 299 {
		return errors.New("PublishPanelResource error" + response.String())
	}
	return nil
}

func (c *SdkClient) PublishPanelResource2(urlValues map[string]interface{}) error {
	//签名问题，暂时用data转json
	req := c.getReq()
	dataJson, err := json.Marshal(urlValues)
	if err != nil {
		return err
	}
	formData := map[string]string{
		"data": string(dataJson),
	}
	response, err := req.SetFormData(formData).SetHeader("Accept", "application/json").Post("/api/thirdparty-cd/k8s-offline/sdk/panel/resource2")
	if err != nil {
		return err
	}
	if response.StatusCode() > 299 {
		return errors.New("PublishPanelResource2 error" + response.String())
	}
	return nil
}

func (c *SdkClient) DeletePanelResource(urlValues map[string]string) error {
	req := c.getReq()
	response, err := req.SetFormData(urlValues).Delete("/api/thirdparty-cd/k8s-offline/sdk/panel/resource")
	if err != nil {
		return err
	}
	if response.StatusCode() > 299 {
		return errors.New("DeletePanelResource error" + response.String())
	}
	return nil
}

func (c *SdkClient) DeletePanelResource2(urlValues map[string]string) error {
	req := c.getReq()
	response, err := req.SetFormData(urlValues).Delete("/api/thirdparty-cd/k8s-offline/sdk/panel/resource2")
	if err != nil {
		return err
	}
	if response.StatusCode() > 299 {
		return errors.New("DeletePanelResource error" + response.String())
	}
	return nil
}

func (c *SdkClient) GenCert(id string) (*LicenseOrder, error) {
	result := &LicenseOrder{}
	errResult := &ConsoleError{}
	req := c.getReq()
	res, err := req.SetResult(result).SetError(errResult).Post("/api/thirdparty-cd/k8s-offline/sdk/license/" + id + "/gentls-by-sdk")
	if err != nil {
		return nil, err
	}
	if res.StatusCode() > 399 {
		// slog.Error("")
		slog.Error("CreateProductOrder", "error", errResult)
		return nil, errResult
	}
	return result, nil

}

func (c *SdkClient) PrepareProduct2() (*GoodsProduct, error) {
	result := &GoodsProduct{}
	req := c.getReq()
	response, err := req.SetResult(result).Post("/api/thirdparty-cd/k8s-offline/sdk/prepare2")
	if err != nil {
		return nil, err
	}
	if response.StatusCode() > 299 {
		slog.Warn("sdk prepare product error", "statusCode", response.StatusCode(), "response", response.String())
		return nil, errors.New("prepare product error" + response.String())
	}
	return result, err
}

func (c *SdkClient) Get(result interface{}, url string, params map[string]string) (interface{}, error) {
	req := c.getReq()
	errResult := &ConsoleError{}
	if params != nil {
		req.SetQueryParams(params)
	}
	response, err := req.SetResult(result).SetError(errResult).Get(url)
	if err != nil {
		return nil, err
	}
	if response.StatusCode() > 299 {
		slog.Warn("sdk get error", "statusCode", response.StatusCode(), "response", response.String())
		return nil, errors.New("sdk get error" + response.String())
	}
	return result, err
}
func (c *SdkClient) Put(result interface{}, url string, params map[string]string) (interface{}, error) {
	req := c.getReq()
	errResult := &ConsoleError{}
	if params != nil {
		req.SetFormData(params)
	}
	response, err := req.SetResult(result).SetError(errResult).Put(url)
	if err != nil {
		return nil, err
	}
	if response.StatusCode() > 299 {
		slog.Warn("sdk get error", "statusCode", response.StatusCode(), "response", response.String())
		return nil, errors.New("sdk get error" + response.String())
	}
	return result, err
}

func (c *SdkClient) Post(result interface{}, url string, params map[string]string) (interface{}, error) {
	req := c.getReq()
	errResult := &ConsoleError{}
	if params != nil {
		req.SetFormData(params)
	}
	response, err := req.SetResult(result).SetError(errResult).Post(url)
	if err != nil {
		slog.Warn("sdk get error", "statusCode", response.StatusCode(), "response", response.String())
		return nil, err
	}

	if response.StatusCode() > 299 {
		slog.Warn("sdk get error", "statusCode", response.StatusCode(), "response", response.String())
		return nil, errors.New("sdk get error" + response.String())
	}
	return result, err
}

func (c *SdkClient) GetCoupon(code string) (*Coupon, error) {
	coupon := &Coupon{}
	_, err := c.Post(coupon, "/api/thirdparty-cd/k8s-offline/sdk/coupon/"+code+"/info", nil)
	return coupon, err
}

func (c *SdkClient) UpdateCoupon(code string, status string, sn string) error {
	// coupon := &CouponCode{}
	params := map[string]string{
		"status": status,
		// "sn":     sn,
	}
	if sn != "" {
		params["sn"] = sn
	}
	_, err := c.Put(nil, "/api/thirdparty-cd/k8s-offline/sdk/coupon/"+code, params)
	return err
}

// 签名问题 需要用Post
func (c *SdkClient) FindLastPaidOrder(clusterId string, k3kName string) (*LastPaidOrder, error) {
	order := &LastPaidOrder{}
	params := map[string]string{
		"clusterId": clusterId,
		"k3kName":   k3kName,
	}
	_, err := c.Post(order, "api/thirdparty-cd/k8s-offline/sdk/panel/lastpaidorder", params)
	return order, err
}

// 返回数据不能用下划线 ????
func (c *SdkClient) FindLastReturnOrder(clusterId string, k3kName string) (*LastReturnOrder, error) {
	order := &LastReturnOrder{}
	params := map[string]string{
		"clusterId": clusterId,
		"k3kName":   k3kName,
	}
	_, err := c.Post(order, "api/thirdparty-cd/k8s-offline/sdk/panel/lastreturnorder", params)
	return order, err
}

func (c *SdkClient) FindK3kOrder(k3kName string, orderSn string) (*K3kOrder, error) {
	order := &K3kOrder{}
	params := map[string]string{
		// "clusterId": clusterId,
		"k3kName": k3kName,
		"orderSn": orderSn,
	}
	_, err := c.Post(order, "/api/thirdparty-cd/k8s-offline/sdk/panel/order", params)
	return order, err
}

func (c *SdkClient) ReturnOrderFinish(k3kName string, sn string) (*LastReturnOrder, error) {
	order := &LastReturnOrder{}
	params := map[string]string{
		// "clusterId": clusterId,
		"k3kName": k3kName,
		"orderSn": sn,
	}
	_, err := c.Put(order, "/api/thirdparty-cd/k8s-offline/sdk/panel/lastreturnorder", params)
	return order, err
}

func (c *SdkClient) CreatePanelOrder(urlValues url.Values) (*PayResult, error) {
	result := &PayResult{}
	response, err := c.getReq().SetFormDataFromValues(urlValues).SetResult(result).Post("/api/thirdparty-cd/k8s-offline/sdk/panel/create-order")
	if err != nil {
		return nil, err
	}
	if response.StatusCode() > 299 {
		slog.Warn("sdk create panel order CreatePanelOrder error", "statusCode", response.StatusCode(), "response", response.String())
		return nil, errors.New("CreatePanelOrder error" + response.String())
	}
	return result, err
}

// 直接代理请求 go不写中间代码 判断只有founder能代理
func (c *SdkClient) Proxy(path string, rawQuery string) (*httputil.ReverseProxy, error) {
	uri, err := url.Parse(c.Client.GetHttpClient().BaseURL)
	if err != nil {
		// c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		slog.Error("Error building kubeAPIURL: %v", "err", err)
		return nil, err
	}
	proxy := httputil.NewSingleHostReverseProxy(uri)

	// 修改 Director 来在请求转发前添加签名
	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		// 先执行原始的 Director
		if originalDirector != nil {
			originalDirector(req)
		}
		req.Host = uri.Host
		req.URL.Scheme = uri.Scheme
		req.URL.Host = uri.Host
		req.URL.Path = path
		if rawQuery != "" {
			// req.URL.RawQuery = rawQuery
		}
		// req.URL.RawQuery =

		// 直接对 http.Request 添加签名
		err = c.addSignatureToHttpRequest(req)
		if err != nil {
			slog.Error("Error adding signature: %v", "err", err)
			return
		}
	}

	return proxy, nil
}
