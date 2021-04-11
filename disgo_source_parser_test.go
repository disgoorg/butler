package main

import (
	"log"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_searchSource(t *testing.T) {
	err := loadPackages("../disgo")
	assert.NoError(t, err)

	result := findInPackages("api", "Cache")
	log.Printf("%+v", result)
}
