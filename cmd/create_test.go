// Copyright 2021-2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package cmd

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"io"
	"os"
	"path/filepath"
	"testing"
)

func TestCreateRecord(t *testing.T) {
	tempFilePath := filepath.Join(t.TempDir(), "out.cbor")
	cmd := NewCreateCmd()

	args := []string{"--data=../testdata/cbor-in-1.cbor", "--output=" + tempFilePath}
	cmd.SetArgs(args)

	err := cmd.Execute()
	assert.NoError(t, err)

	actual, err := os.ReadFile(tempFilePath)
	assert.NoError(t, err)
	expected, err := os.ReadFile("../testdata/expected-record.cbor")
	assert.NoError(t, err)

	assert.Equal(t, expected, actual, "Output record did not have expected contents")
}

func TestCreateCollection(t *testing.T) {
	tempFilePath := filepath.Join(t.TempDir(), "out.cbor")
	cmd := NewCreateCmd()

	args := []string{"--data=../testdata/cbor-in-1.cbor", "--label=cbor1", "--data=../testdata/cbor-in-2.cbor", "--label=cbor2", "--output=" + tempFilePath}
	cmd.SetArgs(args)

	err := cmd.Execute()
	assert.NoError(t, err)

	actual, err := os.ReadFile(tempFilePath)
	assert.NoError(t, err)
	expected, err := os.ReadFile("../testdata/expected-collection.cbor")
	assert.NoError(t, err)

	assert.Equal(t, expected, actual, "Output collection did not have expected contents")
}

// Output to stdout
func TestCreateRecord_Stdout(t *testing.T) {
	old := os.Stdout
	r, w, err := os.Pipe()
	assert.NoError(t, err)

	cmd := NewCreateCmd()

	args := []string{"--data=../testdata/cbor-in-1.cbor"}
	cmd.SetArgs(args)

	os.Stdout = w

	actualC := make(chan []byte)
	go func() {
		var buf bytes.Buffer
		io.Copy(&buf, r)
		actualC <- buf.Bytes()
	}()

	err = cmd.Execute()
	assert.NoError(t, err)

	w.Close()
	os.Stdout = old

	actual := <-actualC // why?

	expected, err := os.ReadFile("../testdata/expected-record.cbor")
	assert.NoError(t, err)

	assert.Equal(t, expected, actual, "Output record did not have expected contents")
}

func TestCreateCollection_Stdout(t *testing.T) {
	old := os.Stdout
	r, w, err := os.Pipe()
	assert.NoError(t, err)

	cmd := NewCreateCmd()

	args := []string{"--data=../testdata/cbor-in-1.cbor", "--label=cbor1", "--data=../testdata/cbor-in-2.cbor", "--label=cbor2"}
	cmd.SetArgs(args)

	os.Stdout = w

	actualC := make(chan []byte)
	go func() {
		var buf bytes.Buffer
		io.Copy(&buf, r)
		actualC <- buf.Bytes()
	}()

	err = cmd.Execute()
	assert.NoError(t, err)

	w.Close()
	os.Stdout = old
	actual := <-actualC

	expected, err := os.ReadFile("../testdata/expected-collection.cbor")
	assert.NoError(t, err)

	assert.Equal(t, expected, actual, "Output collection did not have expected contents")
}
