package initialize

import "context"

type Authentication struct {
	Token string
}

func (a *Authentication) GetRequestMetadata(context.Context, ...string) (map[string]string, error) {

	return map[string]string{"token": a.Token}, nil
}
func (a *Authentication) RequireTransportSecurity() bool {
	return false
}
