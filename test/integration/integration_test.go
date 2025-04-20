// Package integration это интеграционное тестирование
package integration

import (
	"testing"
	"time"
)

func TestFullFlow(t *testing.T) {
	time.Sleep(time.Second * 2) // нужно, чтобы дождаться поднятия БД в докере
	testApp := &TestAppModule{}
	testApp.Invoke(t)
}
