package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestServer_TableName(t *testing.T) {
	model := Server{}
	assert.Equal(t, "net_server", model.TableName())
}
