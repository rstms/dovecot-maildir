package cmd

import (
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestUncompressFiles(t *testing.T) {
	TestInit(t)
	viper.Set("verbose", true)
	err := UncompressMaildirFiles([]string{"testdata/Maildir"})
	require.Nil(t, err)
}
