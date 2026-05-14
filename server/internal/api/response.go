package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// R 统一响应格式
type R struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data any    `json:"data,omitempty"`
}

func OK(c *gin.Context, data any) {
	c.JSON(http.StatusOK, R{Code: 0, Msg: "ok", Data: data})
}

func OKMsg(c *gin.Context, msg string) {
	c.JSON(http.StatusOK, R{Code: 0, Msg: msg})
}

func Fail(c *gin.Context, code int, msg string) {
	c.JSON(http.StatusOK, R{Code: code, Msg: msg})
}

func FailErr(c *gin.Context, err error) {
	c.JSON(http.StatusOK, R{Code: -1, Msg: err.Error()})
}

func Unauthorized(c *gin.Context, msg string) {
	c.JSON(http.StatusUnauthorized, R{Code: 401, Msg: msg})
}

func pageParams(c *gin.Context) (int, int) {
	page := 1
	size := 20
	if v := c.Query("page"); v != "" {
		if p := atoi(v); p > 0 {
			page = p
		}
	}
	if v := c.Query("size"); v != "" {
		if s := atoi(v); s > 0 && s <= 100 {
			size = s
		}
	}
	return page, size
}

func atoi(s string) int {
	n := 0
	for _, c := range s {
		if c >= '0' && c <= '9' {
			n = n*10 + int(c-'0')
		} else {
			return 0
		}
	}
	return n
}

func atoui64(s string) uint64 {
	return uint64(atoi(s))
}
