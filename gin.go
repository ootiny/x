package x

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type HttpResponse struct {
	Data    any    `json:"data,omitempty"`    // Data is the data of the response
	Code    int    `json:"code,omitempty"`    // Code equals 0 means no error, otherwise means error
	Message string `json:"message,omitempty"` // Message is the detail of the error or the success message
}

func GinReturnOK(c *gin.Context, data any) {
	c.JSON(http.StatusOK, HttpResponse{
		Code: 0,
		Data: data,
	})
}

func GinReturnErrorf(c *gin.Context, format string, args ...any) {
	c.JSON(http.StatusOK, HttpResponse{
		Code:    500,
		Message: Sprintf(format, args...),
	})
}

func GinReturnError(c *gin.Context, err error) {
	c.JSON(http.StatusOK, HttpResponse{
		Code:    500,
		Message: err.Error(),
	})
}

func GinReturnResponse(c *gin.Context, response HttpResponse) {
	c.JSON(http.StatusOK, response)
}
