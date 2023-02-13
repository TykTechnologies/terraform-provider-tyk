package tyk

func NewGenericHttpError(body string) GenericHttpError {
	return GenericHttpError{Body: body}
}

type GenericHttpError struct {
	Body string
}

func (g GenericHttpError) Error() string {
	return g.Body
}
