package bitbucket

import (
	"errors"
	"fmt"

	"github.com/go-resty/resty/v2"
	"github.com/tidwall/gjson"
)

// bitbucketIterator allows iterating over
// a paged response of a Bitbucket API call
type bitbucketIterator[T any] struct {
	Client     *BitbucketCloudClient
	RequestURL string
	Parse      func(key, value gjson.Result) (T, error)
	hasNext    bool
	nextURL    string
}

// newBitbucketIteratorOptions is the options for creating a new bitbucket iterator
type newBitbucketIteratorOptions[T any] struct {
	// Client is the bitbucket client
	Client *BitbucketCloudClient
	// RequestURL is the request URL
	RequestURL string
	// Parse is the function to parse the response
	Parse func(key, value gjson.Result) (T, error)
}

// newBitbucketIterator creates a new bitbucket iterator
func newBitbucketIterator[T any](options *newBitbucketIteratorOptions[T]) *bitbucketIterator[T] {
	return &bitbucketIterator[T]{
		Client:     options.Client,
		RequestURL: options.RequestURL,
		Parse:      options.Parse,
		hasNext:    true,
		nextURL:    "",
	}
}

func (i *bitbucketIterator[T]) HasNext() bool {
	return i.hasNext
}

// GetAll returns a list values from all pages
func (i *bitbucketIterator[T]) GetAll() ([]T, error) {
	// FIXME: Return a channel instead?
	result := []T{}
	for i.HasNext() {
		lists, err := i.Next()
		if err != nil {
			return nil, err
		}

		result = append(result, lists...)
	}

	return result, nil
}

func (i *bitbucketIterator[T]) sendRequest(
	request *resty.Request,
) ([]T, error) {
	r, err := request.Send()
	if err != nil {
		return nil, err
	}
	if r.IsError() {
		return nil, errors.New(string(r.Body()))
	}
	parsed := gjson.ParseBytes(r.Body())

	i.nextURL = parsed.Get("next").String()
	if i.nextURL == "" {
		i.hasNext = false
	}

	return i.parse(parsed)
}

func (i *bitbucketIterator[T]) parse(parsed gjson.Result) ([]T, error) {
	list := []T{}

	result := parsed.Get("values")
	result.ForEach(func(key, value gjson.Result) bool {
		obj, err := i.Parse(key, value)
		if err != nil {
			return false
		}

		list = append(list, obj)
		return true
	})

	return list, nil
}

func (i *bitbucketIterator[T]) doInitialCall() ([]T, error) {
	const pageLength = 50
	url := i.RequestURL

	r := resty.New().R().
		SetBasicAuth(i.Client.username, i.Client.password).
		SetQueryParam("pagelen", fmt.Sprint(pageLength)).
		SetError(bbError{})
	r.Method = "GET"
	r.URL = url

	return i.sendRequest(r)
}

func (i *bitbucketIterator[T]) doNextCall() ([]T, error) {
	r := resty.New().R().
		SetBasicAuth(i.Client.username, i.Client.password).
		SetError(bbError{})
	r.URL = i.nextURL

	return i.sendRequest(r)
}

func (i *bitbucketIterator[T]) Next() ([]T, error) {
	if !i.hasNext {
		return nil, nil
	}

	if i.nextURL == "" {
		return i.doInitialCall()
	} else {
		return i.doNextCall()
	}
}
