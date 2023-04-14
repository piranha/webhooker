// (c) 2013 Alexander Solovyov

package main

import (
	"github.com/stretchr/testify/mock"
	"testing"
)

type MockedRule struct {
	mock.Mock
	Pattern string
}

func (r *MockedRule) Match(path string) bool {
	r.Mock.Called(path)
	return path == r.Pattern
}

func (r *MockedRule) String() string {
	return r.Pattern
}

func (r *MockedRule) Run(data Payload) (string, error) {
	args := r.Mock.Called(data)
	return "", args.Error(0)
}

func TestRunIsCalled(t *testing.T) {
	data := &GithubPayload{
		Ref: "refs/heads/main",
		Repository: GithubRepo{
			Name:     "webhooker",
			FullName: "piranha/webhooker",
		},
	}

	wrong := &MockedRule{
		Pattern: "nothing",
	}
	right := &MockedRule{
		Pattern: "piranha/webhooker:main",
	}

	wrong.On("Match", GetPath(data)).Return(false)
	right.On("Match", GetPath(data)).Return(true)
	right.On("Run", data).Return(nil)

	c := &Config{wrong, right}
	_, err := c.ExecutePayload(data)
	if err != nil {
		t.Error("ExecutePayload should not return an error")
	}

	wrong.Mock.AssertExpectations(t)
	right.Mock.AssertExpectations(t)
}
