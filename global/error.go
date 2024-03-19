package global

import "errors"

var (
	GatewayIdMismatchError   = errors.New("The gateway ID corresponding to the host does not match the gateway ID of the machine room.")
	GroupNameRepeatError     = errors.New("Duplicate group name error")
	GroupIDDeletedBoundError = errors.New("The deleted ID is bound with host information")
)
