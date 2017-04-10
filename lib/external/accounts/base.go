package accounts

import (
	"fmt"
	"net/http"
)

const domain = "https://dev.account.ely.by"

var Client = &http.Client{}

type UnauthorizedResponse struct {}

func (err UnauthorizedResponse) Error() string {
	return "Unauthorized response"
}

type ForbiddenResponse struct {}

func (err ForbiddenResponse) Error() string {
	return "Forbidden response"
}

type NotFoundResponse struct {}

func (err NotFoundResponse) Error() string {
	return "Not found"
}

type NotSuccessResponse struct {
	StatusCode int
}

func (err NotSuccessResponse) Error() string {
	return fmt.Sprintf("Response code is \"%d\"", err.StatusCode)
}

func handleResponse(response *http.Response) error {
	switch status := response.StatusCode; status {
	case 200:
		return nil
	case 401:
		return &UnauthorizedResponse{}
	case 403:
		return &ForbiddenResponse{}
	case 404:
		return &NotFoundResponse{}
	default:
		return &NotSuccessResponse{status}
	}
}
