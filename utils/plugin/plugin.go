package plugin

import (
	"github.com/gin-gonic/gin"
)

const (
	OnlyFuncName = "Plugin"
)

// Plugin Plugin Mode Interfacization
type Plugin interface {
	// Register Register Routing
	Register(group *gin.RouterGroup)

	// RouterPath User returns registered route
	RouterPath() string
}
