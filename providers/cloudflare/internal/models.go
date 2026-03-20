package internal

import (
	"errors"
	"fmt"
)

type baseRespMessage struct {
	Code             int    `json:"code"`
	Message          string `json:"message"`
	DocumentationUrl string `json:"documentation_url"`
	// ignore other fields
}

type baseRespResultInfo struct {
	Count      int `json:"count"`
	Page       int `json:"page"`
	PerPage    int `json:"per_page"`
	TotalCount int `json:"total_count"`
	TotalPages int `json:"total_pages"`
}

type apiResponse[T any] struct {
	Success    bool               `json:"success"`
	Results    []T                `json:"results"`
	Errors     []baseRespMessage  `json:"errors"`
	Messages   []baseRespMessage  `json:"messages"`
	ResultInfo baseRespResultInfo `json:"result_info"`
}

func (ap *apiResponse[T]) Err() error {
	var err error
	for i := 0; i < len(ap.Errors); i++ {
		E := ap.Errors[i]
		err = errors.Join(err, fmt.Errorf("remote error, code=%d,message=%s,doc=%s", E.Code, E.Message, E.DocumentationUrl))
	}
	return err
}

type Zone struct {
	ID   string `json:"id"`
	Name string `json:"name"`

	// other doesn't needed fields are ignored
}
