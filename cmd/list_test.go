package cmd

import (
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestListAll(t *testing.T) {
	TestInit(t)
	viper.Set("all", true)
	files, err := ListMaildirFiles("testdata/Maildir")
	require.Nil(t, err)
	require.IsType(t, &[]string{}, files)
	require.NotEmpty(t, *files)
}

func TestListCompressed(t *testing.T) {
	TestInit(t)
	files, err := ListMaildirFiles("testdata/Maildir")
	require.Nil(t, err)
	require.IsType(t, &[]string{}, files)
	require.NotEmpty(t, *files)
}

func TestListUncompressed(t *testing.T) {
	TestInit(t)
	viper.Set("uncompressed", true)
	files, err := ListMaildirFiles("testdata/Maildir")
	require.Nil(t, err)
	require.IsType(t, &[]string{}, files)
	require.NotEmpty(t, *files)
}
