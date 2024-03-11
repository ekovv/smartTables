package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"smartTables/internal/constants"
)

func HandlerErr(c *gin.Context, err error) {
	var UnmarshalTypeError *json.UnmarshalTypeError
	if err != nil {
		switch {
		case errors.Is(err, constants.ErrAlreadyExists):
			err := fmt.Sprintf("already exists %s", err)
			c.JSON(http.StatusNotFound, err)
		case errors.Is(err, constants.ErrInvalidData):
			err := fmt.Sprintf("invalid data %s", err)
			c.JSON(http.StatusNotFound, err)
		case errors.As(err, &UnmarshalTypeError):
			err := fmt.Sprintf("bad json %s", err)
			c.JSON(http.StatusBadRequest, err)
		default:
			c.JSON(http.StatusBadRequest, err)
		}
		return
	}
	c.Status(http.StatusOK)
	return
}
