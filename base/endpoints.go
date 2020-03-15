package base

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-kit/kit/endpoint"
)

var ErrBadRequest = errors.New("bad request")

type Endpoints struct {
	Check       endpoint.Endpoint
	GetDocument endpoint.Endpoint
	PutDocument endpoint.Endpoint
}

func NewServerEndPoints(s Service) Endpoints {
	return Endpoints{
		Check:       MakeCheck(s),
		GetDocument: MakeGetDocument(s),
		PutDocument: MakePutDocument(s),
	}
}

func MakeCheck(s Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		fmt.Println("hi")
		return s.Check(ctx)
	}
}

func MakeGetDocument(s Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		if val, ok := request.(string); ok {
			return s.GetDocument(ctx, val)
		}
		return nil, ErrBadRequest
	}
}

func MakePutDocument(s Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		return s.PutDocument(ctx)
	}
}
