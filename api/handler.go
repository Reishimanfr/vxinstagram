/*
VxInstagram - Blazing fast embedder for instagram posts
Copyright (C) 2025 Bash06

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>.
*/
package api

import (
	"bash06/vxinstagram/flags"
	"bash06/vxinstagram/middleware"
	"net/http"
	"time"

	cache "github.com/chenyahui/gin-cache"
	"github.com/chenyahui/gin-cache/persist"
	sentrygin "github.com/getsentry/sentry-go/gin"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

type Handler struct {
	Db     *gorm.DB
	Router *gin.Engine
}

// Attaches middleware and sets endpoint funcs
func NewHandler(db *gorm.DB) *Handler {
	r := gin.New()

	r.Use(
		gin.Recovery(),
		gin.ErrorLogger(),
		middleware.RateLimiterMiddleware(middleware.NewRateLimiter(5, 10)),
		middleware.CorsMiddleware(),
		sentrygin.New(sentrygin.Options{}),
	)

	r.LoadHTMLGlob("templates/*")

	return &Handler{
		Db:     db,
		Router: r,
	}
}

func (h *Handler) Init() {
	var st persist.CacheStore = persist.NewMemoryStore(time.Minute * 1)
	cacheExpire := time.Minute * time.Duration(*flags.CacheLifetime)

	if *flags.RedisEnable {
		rdb := redis.NewClient(&redis.Options{
			Addr:     *flags.RedisAddr,
			Password: *flags.RedisPasswd,
			DB:       *flags.RedisDB,
		})

		st = persist.NewRedisStore(rdb)
	}

	h.Router.GET("/reel/:id", cache.CacheByRequestURI(st, time.Minute*1), h.ServeVideo)
	h.Router.GET("/reels/:id", cache.CacheByRequestURI(st, cacheExpire), h.ServeVideo)
	h.Router.GET("/p/:id", cache.CacheByRequestURI(st, cacheExpire), h.ServeVideo)
	h.Router.GET("/favicon.ico", func(ctx *gin.Context) { ctx.Status(http.StatusOK) })

	h.Router.GET("/", func(ctx *gin.Context) {
		ctx.Redirect(http.StatusPermanentRedirect, "https://github.com/Reishimanfr/vxinstagram?tab=readme-ov-file#how-to-use")
	})

	h.Router.GET("/share/:id", cache.CacheByRequestURI(st, cacheExpire), h.FollowShare)

}
