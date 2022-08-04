package routerUtils

import (
	"github.com/gin-gonic/gin"
)

func GenerateSuccessResponse(data gin.H) gin.H {
	return gin.H{
		"data":    data,
		"success": true,
		"message": "OK",
	}
}

func GenerateErrorResponse(err error) gin.H {
	return gin.H{
		"data":    gin.H{},
		"success": false,
		"message": err.Error(),
	}
}

func SendSuccessResponse(ctx *gin.Context, statusCode int, data gin.H) {
	ctx.IndentedJSON(statusCode, GenerateSuccessResponse(data))
}

func SendErrorResponse(ctx *gin.Context, statusCode int, err error) {
	ctx.IndentedJSON(statusCode, GenerateErrorResponse(err))
}
