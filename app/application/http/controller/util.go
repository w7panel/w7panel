package controller

import (
	// "archive/zip"

	"encoding/base64"
	"encoding/json"
	"log"
	"log/slog"
	"net"
	"strings"
	"time"

	"gitee.com/we7coreteam/k8s-offline/common/helper"
	"github.com/gin-gonic/gin"
	pinyin "github.com/mozillazg/go-pinyin"
	"github.com/we7coreteam/w7-rangine-go/v2/pkg/support/facade"

	// "github.com/we7coreteam/w7-rangine-go/v2/src/core/helper"
	"github.com/we7coreteam/w7-rangine-go/v2/src/http/controller"
	"github.com/wenlng/go-captcha-assets/resources/images"
	"github.com/wenlng/go-captcha-assets/resources/tiles"
	"github.com/wenlng/go-captcha/v2/slide"
)

type Util struct {
	controller.Abstract
}

var slideCapt slide.Captcha

func init() {
	builder := slide.NewBuilder(
		//slide.WithGenGraphNumber(2),
		slide.WithEnableGraphVerticalRandom(true),
	)

	// background images
	imgs, err := images.GetImages()
	if err != nil {
		slog.Warn("img err", "error", err)
	}

	graphs, err := tiles.GetTiles()
	if err != nil {
		slog.Warn("graphs err", "error", err)
	}

	var newGraphs = make([]*slide.GraphImage, 0, len(graphs))
	for i := 0; i < len(graphs); i++ {
		graph := graphs[i]
		newGraphs = append(newGraphs, &slide.GraphImage{
			OverlayImage: graph.OverlayImage,
			MaskImage:    graph.MaskImage,
			ShadowImage:  graph.ShadowImage,
		})
	}

	// set resources
	builder.SetResources(
		slide.WithGraphImages(newGraphs),
		slide.WithBackgrounds(imgs),
	)

	slideCapt = builder.Make()
}

func (self Util) Pinyin(http *gin.Context) {

	type ParamsValidate struct {
		Words string `form:"words" binding:"required"`
	}
	params := ParamsValidate{}
	if !self.Validate(http, &params) {
		return
	}
	result := pinyin.Convert(params.Words, nil)
	// result = strings.ToLower(result)
	self.JsonResponse(http, result, nil, 200)
}

func (self Util) DnsIp(http *gin.Context) {

	type ParamsValidate struct {
		Domain string `form:"domain" binding:"required"`
	}
	params := ParamsValidate{}
	if !self.Validate(http, &params) {
		return
	}
	result, err := net.LookupIP(params.Domain)
	if err != nil {
		self.JsonResponse(http, []string{}, nil, 200)
	}
	self.JsonResponse(http, result, nil, 200)
}

func (self Util) DnsCName(http *gin.Context) {

	type ParamsValidate struct {
		Domain string `form:"domain" binding:"required"`
	}
	params := ParamsValidate{}
	if !self.Validate(http, &params) {
		return
	}
	result, err := net.LookupCNAME(params.Domain)

	if err != nil {
		self.JsonResponse(http, []string{}, nil, 200)
		return
	}
	// clear end .
	result = strings.TrimRight(result, ".")
	self.JsonResponse(http, []string{result}, nil, 200)
}

func (self Util) MyIp(http *gin.Context) {
	result, err := helper.MyIp()
	if err != nil {
		self.JsonResponse(http, []string{}, nil, 200)
	}
	result2 := map[string]string{"ip": result}
	self.JsonResponse(http, result2, nil, 200)
}

func (self Util) Captcha(http *gin.Context) {
	captData, err := slideCapt.Generate()
	if err != nil {
		log.Fatalln(err)
	}

	blockData := captData.GetData()
	if blockData == nil {
		self.JsonResponseWithServerError(http, err)
		return
	}

	var masterImageBase64, tileImageBase64 string
	masterImageBase64 = captData.GetMasterImage().ToBase64()
	if err != nil {
		self.JsonResponseWithServerError(http, err)
		return
	}

	tileImageBase64 = captData.GetTileImage().ToBase64()
	if err != nil {
		self.JsonResponseWithServerError(http, err)
		return
	}

	dotsByte, _ := json.Marshal(blockData)
	if len(dotsByte) == 0 {
		self.JsonResponseWithServerError(http, err)
		return
	}
	pass := facade.GetConfig().GetString("app.aeskey")
	encrypt, err := helper.AES_encrypt(dotsByte, []byte(pass))
	if err != nil {
		self.JsonResponseWithServerError(http, err)
		return
	}
	key := base64.URLEncoding.EncodeToString(encrypt)
	// key := helper.StringToMD5(string(dotsByte))

	bt := map[string]interface{}{
		"code":         0,
		"captcha_key":  key,
		"image_base64": masterImageBase64,
		"tile_base64":  tileImageBase64,
		"tile_width":   blockData.Width,
		"tile_height":  blockData.Height,
		"tile_x":       blockData.TileX,
		"tile_y":       blockData.TileY,
	}
	self.JsonResponse(http, bt, nil, 200)

}

/*
*

	验证是否可以连接数据库
*/
func (self Util) DbConnTest(http *gin.Context) {
	type ParamsValidate struct {
		Dsn string `form:"dsn" binding:"required"`
	}
	params := ParamsValidate{}
	if !self.Validate(http, &params) {
		return
	}
	ok, err := helper.DBConnTest(params.Dsn)
	if err != nil {
		self.JsonResponseWithoutError(http, gin.H{"canConnect": false, "msg": err.Error()})
		return
	}
	self.JsonResponseWithoutError(http, gin.H{"canConnect": ok})
}

func (self Util) PintEtcd(ginctx *gin.Context) {
	type ParamsValidate struct {
		Url string `form:"url" binding:"required"`
	}
	params := ParamsValidate{}
	if !self.Validate(ginctx, &params) {
		return
	}
	//判断url尾部是否/结尾
	if !strings.HasSuffix(params.Url, "/") {
		params.Url += "/"
	}
	response, err := helper.RetryHttpClient().SetTimeout(time.Second * 5).GetClient().Get(params.Url + "health")
	if err != nil {
		self.JsonResponseWithoutError(ginctx, gin.H{"canConnect": false, "msg": err.Error()})
		return
	}
	self.JsonResponseWithoutError(ginctx, gin.H{"canConnect": response.StatusCode == 200})

}

// 检测验证码是否正确
func (self Util) VerifyCaptcha(ginctx *gin.Context) {
	type ParamsValidate struct {
		Point string `form:"point" binding:"required"`
		Key   string `form:"key" binding:"required"`
	}

	params := ParamsValidate{}
	if !self.Validate(ginctx, &params) {
		return
	}
	if facade.Config.GetBool("captcha.enabled") {
		err := helper.VerifyCaptcha(params.Point, params.Key, false)
		if err != nil {
			self.JsonResponseWithoutError(ginctx, gin.H{"ok": false, "msg": err.Error()})
			return
		}
	}
	self.JsonResponseWithoutError(ginctx, gin.H{"ok": true})

}
