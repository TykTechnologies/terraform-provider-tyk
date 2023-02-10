package tyk

import "fmt"

type GenericFlagError struct {
	FlagName string
}

func (g *GenericFlagError) Error() string {
	return fmt.Sprintf("error getting  %s flag", g.FlagName)
}

func NewGenericFlagError(flagName string) GenericFlagError {
	return GenericFlagError{FlagName: flagName}
}

func NewGenericHttpError(body string) GenericHttpError {
	return GenericHttpError{Body: body}
}

type GenericHttpError struct {
	Body string
}

func (g GenericHttpError) Error() string {
	return g.Body
}
