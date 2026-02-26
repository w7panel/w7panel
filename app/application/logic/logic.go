package logic

import (
	"github.com/redis/go-redis/v9"
	"github.com/we7coreteam/w7-rangine-go/v2/pkg/support/facade"
	"gorm.io/gorm"
)

type logic struct {
}

func (l logic) GetDefaultDb() *gorm.DB {
	db, _ := facade.GetDbFactory().Channel("default")
	return db
}

func (l logic) GetDefaultRedis() redis.Cmdable {
	redis, _ := facade.GetRedisFactory().Channel("default")

	return redis
}
