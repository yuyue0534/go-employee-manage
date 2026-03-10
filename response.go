package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type RespBody struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

type PagedData struct {
	Items any   `json:"items"`
	Total int64 `json:"total"`
	Page  int   `json:"page"`
	Size  int   `json:"size"`
}

func respOK(c *gin.Context, data any) {
	c.JSON(http.StatusOK, RespBody{Code: 0, Message: "ok", Data: data})
}

func respCreated(c *gin.Context, data any) {
	c.JSON(http.StatusCreated, RespBody{Code: 0, Message: "created", Data: data})
}

func respPaged(c *gin.Context, items any, total int64, page, size int) {
	c.JSON(http.StatusOK, RespBody{
		Code: 0, Message: "ok",
		Data: PagedData{Items: items, Total: total, Page: page, Size: size},
	})
}

func respBadRequest(c *gin.Context, msg string) {
	c.AbortWithStatusJSON(http.StatusBadRequest, RespBody{Code: 400, Message: msg})
}

func respNotFound(c *gin.Context, msg string) {
	c.AbortWithStatusJSON(http.StatusNotFound, RespBody{Code: 404, Message: msg})
}

func respConflict(c *gin.Context, msg string) {
	c.AbortWithStatusJSON(http.StatusConflict, RespBody{Code: 409, Message: msg})
}

func respError(c *gin.Context, msg string) {
	c.AbortWithStatusJSON(http.StatusInternalServerError, RespBody{Code: 500, Message: msg})
}
