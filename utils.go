package pure

import (
	"context"
	"mime/multipart"
)

const (
	methodContextKey    = ContextKey("HTTPMethod")
	multipartContextKey = ContextKey(MultipartFormData)
)

// HTTPMethod returns the HTTP method of request associated in Context.
func HTTPMethod(ctx context.Context) string {
	method, ok := ctx.Value(methodContextKey).(string)
	if !ok {
		method = ""
	}
	return method
}

// MultipartForm returns the multipart form data associated in Context.
func MultipartForm(ctx context.Context) *multipart.Form {
	mForm, ok := ctx.Value(multipartContextKey).(*multipart.Form)
	if !ok {
		mForm = nil
	}
	return mForm
}
