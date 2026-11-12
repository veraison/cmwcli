// Copyright 2021-2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package cmd

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/veraison/cmw"
	"os"
)

var (
	cmwCreateFiles      []string
	cmwCreateLabels     []string
	cmwCreateOutputFile string
	cmwCreateCmwc_t     string
)

// createCmd represents the create command
var createCmd = NewCreateCmd()

// NewCreateCmd builds the create command
func NewCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a Conceptual Message Wrapper from (CBOR only) evidence file(s)",
		Long: `Create a Conceptual Message Wrapper from (CBOR only) evidence file(s)
	If only one evidence file are input, then a CMW with a record is created.
	If multiple evidence files are input, then a CMW with a collection of records is created.
	For example:

	cmwcli create --data evidence.cbor --label "evidence" --data evidence2.cbor --label "evidence2" --cmwc_t "https://example.com" --output out.cbor`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := checkCmwCreateArgs(); err != nil {
				return err
			}

			var cmwOutput cmw.CMW
			// Check how many files there are
			if len(cmwCreateFiles) > 1 {
				cmwOutputRef, err := cmw.NewCollection(cmwCreateCmwc_t)
				if err != nil {
					return fmt.Errorf("Error creating collection: %s", err)
				}
				cmwOutput = *cmwOutputRef

				for i, d := range cmwCreateFiles {
					file, err := os.ReadFile(d)
					if err != nil {
						return fmt.Errorf("Error accessing cbor data (%s) in file: %s", d, err)
					}

					newCMWItem, err := cmw.NewMonad("application/eat+cbor", file, 4)
					err = cmwOutput.AddCollectionItem(cmwCreateLabels[i], newCMWItem)
					if err != nil {
						return fmt.Errorf("Error adding evidence from file (%s) to CMW collection: %s", d, err)
					}
				}
			} else {
				file, err := os.ReadFile(cmwCreateFiles[0])
				if err != nil {
					return fmt.Errorf("Error accessing cbor data in file %s: %s", cmwCreateFiles[0], err)
				}
				cmwOutputRef, err := cmw.NewMonad("application/eat+cbor", file, 4)
				if err != nil {
					return fmt.Errorf("Error creating CMW Record from file %s: %s", cmwCreateFiles[0], err)
				}
				cmwOutput = *cmwOutputRef
			}

			// Write to output file
			fmt.Println(cmwOutput)

			cmwBytes, err := cmwOutput.MarshalCBOR()
			if err != nil {
				return errors.New("Error marshalling CMW to a CBOR file. Maybe check the input file is valid CBOR?")
			}

			err = os.WriteFile(cmwCreateOutputFile, cmwBytes, 0666)
			if err != nil {
				return errors.New("Error writing to file")
			}
			return nil
		},
	}
	cmd.Flags().StringArrayVarP(&cmwCreateFiles, "data", "d", []string{}, "Location of input data")
	cmd.Flags().StringArrayVarP(&cmwCreateLabels, "label", "l", []string{}, "Label for input data")
	cmd.Flags().StringVarP(&cmwCreateOutputFile, "output", "o", "", "Location of output file")
	cmd.Flags().StringVarP(&cmwCreateCmwc_t, "cmwc_t", "c", "", "The __cmwc_t of the parent collection")
	return cmd
}
func checkCmwCreateArgs() error {
	if len(cmwCreateFiles) == 0 {
		return errors.New("No data inputted")
	}

	if len(cmwCreateFiles) > 1 && len(cmwCreateFiles) != len(cmwCreateLabels) {
		return fmt.Errorf("Non-equal amount of files and labels input. Given %d files, but %d labels.", len(cmwCreateFiles), len(cmwCreateLabels))
	}

	if cmwCreateOutputFile == "" {
		return errors.New("No output file specified")
	}

	return nil
}

func init() {
	rootCmd.AddCommand(createCmd)
}
