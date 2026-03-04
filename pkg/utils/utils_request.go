package utils

import (
	"errors"
	"helia/internal/common"
	"strconv"

	"github.com/gin-gonic/gin"
)

// GetInt64FromParameterRequest extracts the ID from the request URL path.
func GetInt64FromParameterRequest(c *gin.Context, key string) (int64, error) {
	idStr := c.Param(key)
	if idStr == "" {
		return 0, errors.New(common.ErrMsgGetIDFromURL)
	}
	return strconv.ParseInt(idStr, 10, 64)
}

// GetInt64FromQueryRequest extracts the ID from the request URL query parameters.
func GetInt64FromQueryRequest(c *gin.Context, key string) (int64, error) {
	idStr := c.Query(key)
	if idStr == "" {
		return 0, errors.New(common.ErrMsgGetIDFromURL)
	}
	return strconv.ParseInt(idStr, 10, 64)
}

// GetStringFromQueryRequest extracts the string value from the request URL query parameters.
func GetStringFromQueryRequest(c *gin.Context, key string) (string, error) {
	idStr := c.Query(key)
	if idStr == "" {
		return "", errors.New(common.ErrMsgGetIDFromURL)
	}
	return idStr, nil
}
