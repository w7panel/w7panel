package middleware

import (
	"encoding/json"
	"log/slog"

	"gitee.com/we7coreteam/k8s-offline/common/service/k8s"
	"github.com/gin-gonic/gin"
	"github.com/we7coreteam/w7-rangine-go/v2/src/http/middleware"
)

type BindConsole struct {
	middleware.Abstract
}

func (self BindConsole) Process(ctx *gin.Context) {

	ctx.Next()
	if ctx.Writer.Status() == 200 {
		slog.Info("bind console middleware")
		uid := ctx.Writer.Header().Get("uid")
		if uid != "" {
			token := ctx.MustGet("k8s_token").(string)
			k3ktoken := k8s.NewK8sToken(token)
			saName, err := k3ktoken.GetSaName()
			if err != nil {
				return
			}
			sdk := k8s.NewK8sClient().Sdk
			data := map[string]interface{}{
				"metadata": map[string]interface{}{
					"labels": map[string]string{
						"w7.cc/console-id": uid,
					},
				},
			}
			databytes, err := json.Marshal(data)
			if err != nil {
				slog.Error("bind console middleware json marshal error", "error", err)
				return
			}
			//patch namespace Label w7.cc/console-id: uid
			//patch serviceAccount Label w7.cc/console-id: uid
			_, err = sdk.PatchServiceAccount("", saName, databytes)
			if err != nil {
				slog.Error("bind console middleware patch serviceAccount error", "error", err)
				return
			}

		}
	}

}
