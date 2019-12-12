package mocks

import (
	"context"

	"github.com/cage1016/gokitistiok8s/pkg/addsvc/service"
)

// Export for testing
type AddsvcMock struct {
}

func (m AddsvcMock) Sum(ctx context.Context, a int64, b int64) (rs int64, err error) {
	return a + b, err
}

func (m AddsvcMock) Concat(ctx context.Context, a string, b string) (rs string, err error) {
	return a + b, err
}

func NewAddsvcMock() service.AddsvcService {
	return &AddsvcMock{}
}
