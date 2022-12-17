package main

import (
	"github.com/stretchr/testify/suite"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}

func TestUserService(t *testing.T) {
	suite.Run(t, new(UserServiceSuite))
}
