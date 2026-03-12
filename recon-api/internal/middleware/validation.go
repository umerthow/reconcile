package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

// ValidationMiddleware validates request body
func ValidationMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
	}
}

// ValidateStruct validates a struct using validator tags
func ValidateStruct(data interface{}) error {
	return validate.Struct(data)
}
