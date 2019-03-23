package main_test

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"testing"

	main "github.com/m-mizutani/gsuite-log-exporter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func loadConfig() main.Arguments {
	confPath := os.Getenv("TEST_CONFIG_PATH")
	if confPath == "" {
		confPath = "./test.json"
	}

	raw, err := ioutil.ReadFile(confPath)
	if err != nil {
		log.Fatalf("Can not read file: %s", confPath)
	}

	var args main.Arguments
	if err := json.Unmarshal(raw, &args); err != nil {
		log.Fatalf("Can not unmarshal file: %s", confPath)
	}

	return args
}

func Test(t *testing.T) {
	args := loadConfig()
	resp, err := main.Handler(args)
	require.NoError(t, err)
	assert.NotEqual(t, 0, resp.LogCount)
}
