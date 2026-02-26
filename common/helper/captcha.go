package helper

import (
	"encoding/base64"
	osjson "encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/we7coreteam/w7-rangine-go/v2/pkg/support/facade"
	"github.com/wenlng/go-captcha/v2/slide"
)

var verifyKey = make(map[string]bool)

func VerifyCaptcha(point string, captchaKey string, log bool) error {
	if verifyKey[captchaKey] {
		return fmt.Errorf("验证码已使用")
	}
	if log {
		defer func() {
			verifyKey[captchaKey] = true
		}()
	}
	return verifyCaptcha0(point, captchaKey)
}

func verifyCaptcha0(point string, captchaKey string) error {
	aesKey := facade.GetConfig().GetString("app.aesKey")
	if len(captchaKey) == 0 {
		return fmt.Errorf("captcha key is empty")
	}
	d64captchaKey, err := base64.URLEncoding.DecodeString(captchaKey)
	if err != nil {
		return err
	}
	cacheDataByte, err := AES_decrypt(d64captchaKey, []byte(aesKey))
	if err != nil {
		return err
	}
	var dct *slide.Block
	if err := osjson.Unmarshal(cacheDataByte, &dct); err != nil {
		return err
	}
	src := strings.Split(point, ",")
	chkRet := false
	if 2 == len(src) {
		sx, _ := strconv.ParseFloat(fmt.Sprintf("%v", src[0]), 64)
		sy, _ := strconv.ParseFloat(fmt.Sprintf("%v", src[1]), 64)
		chkRet = slide.CheckPoint(int64(sx), int64(sy), int64(dct.X), int64(dct.Y), 4)
	}
	if chkRet {
		return nil
	}
	return fmt.Errorf("验证码错误")

}
