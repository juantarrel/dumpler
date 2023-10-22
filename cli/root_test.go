package cli

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestCreateViperConfig(t *testing.T) {
	config, err := createViperConfig()
	require.NoError(t, err)
	require.NotNil(t, config)
}
