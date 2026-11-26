// Copyright 2021-2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package cmd

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/spf13/cobra"
	"github.com/veraison/cmw"
	"os"
)

var (
	cmwCreateData       []string
	cmwCreateOutputFile string
	cmwCreateCmwc_t     string
)

type data struct {
	label     string
	indicator cmw.Indicator
	filePath  string
}

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

	cmwcli create --data label1:evidence:evidence1.cbor --data label2:evidence:evidence2.cbor --cmwc_t "https://example.com" --output out.cbor`,
		RunE: func(cmd *cobra.Command, args []string) error {
			data, err := parseCmwCreateArgs()
			if err != nil {
				return err
			}

			// Convert all files into records
			recordRefs := []*cmw.CMW{}
			for _, d := range data {
				file, err := os.ReadFile(d.filePath)
				if err != nil {
					return fmt.Errorf("error accessing cbor data (%s) in file: %s", d.filePath, err)
				}
				newRecordRef, err := cmw.NewMonad("application/eat+cbor", file, d.indicator)
				if err != nil {
					return fmt.Errorf("error converting file (%s) into CMW record: %s", d.filePath, err)
				}
				recordRefs = append(recordRefs, newRecordRef)
			}

			// If > 1 record created, add all records to a collection
			var cmwOutput cmw.CMW
			if len(recordRefs) == 1 {
				// Create record
				cmwOutput = *recordRefs[0]
			} else {
				// Create collection
				cmwOutputRef, err := cmw.NewCollection(cmwCreateCmwc_t)
				if err != nil {
					return fmt.Errorf("error creating collection: %s", err)
				}
				cmwOutput = *cmwOutputRef

				// Add all records to collection
				for i, record := range recordRefs {
					err = cmwOutput.AddCollectionItem(data[i].label, record)
					if err != nil {
						return fmt.Errorf("error adding evidence from file (%s) to CMW collection: %s", data[i].filePath, err)
					}
				}
			}

			cmwBytes, err := cmwOutput.MarshalCBOR()
			if err != nil {
				return errors.New("error marshalling CMW to a CBOR file. Maybe check the input file is valid CBOR?")
			}

			if cmwCreateOutputFile == "" {
				// Write to output file
				fmt.Printf("%s", cmwBytes)
			} else {
				// Write to file
				err = os.WriteFile(cmwCreateOutputFile, cmwBytes, 0666)
				if err != nil {
					return errors.New("error writing to file")
				}
			}
			return nil
		},
	}
	cmd.Flags().StringArrayVarP(&cmwCreateData, "data", "d", []string{}, `Input Data, in form label:indicator:file (or indicator:file for a record).
An indicator can take the values: reference-values (r), endorsements (en), evidence (ev), attestation-results (ar), appraisal-policy (ap).`)
	cmd.Flags().StringVarP(&cmwCreateOutputFile, "output", "o", "", "Location of output file")
	cmd.Flags().StringVarP(&cmwCreateCmwc_t, "cmwc_t", "c", "", "The __cmwc_t of the parent collection")
	return cmd

}
func parseCmwCreateArgs() ([]data, error) {
	var createDataStruct []data // Return value

	// Split cmwCreateData in label, indicator, filepath
	createData := make([]([]string), len(cmwCreateData))
	for i, v := range cmwCreateData {
		createData[i] = strings.SplitN(v, ":", 3)
	}

	// Case: creating record
	if len(createData) == 1 {
		if len(createData[0]) != 2 {
			return nil, fmt.Errorf("attempted to create record with improper --data argument. Expected indicator:filePath, instead recieved %d terms: %s", len(createData[0]), createData[0])
		}

		createData[0] = slices.Concat([]string{""}, createData[0]) // Add dummy label of ""
		// Case: creating collection
	} else {
		for _, v := range createData {
			if len(v) != 3 {
				return nil, fmt.Errorf("attempted to create collection with improper aguments. Expected all --data arguments to be in form label:indicator:filePath. Instead, recieved --data argument with %d terms: %s", len(v), v)
			}
		}

	}

	// Create final struct and return
	createDataStruct = make([]data, len(createData))
	for i, v := range createData {
		var ind cmw.Indicator
		switch v[1] {
		case "reference-values", "r":
			ind.Set(cmw.ReferenceValues)
		case "endorsements", "en":
			ind.Set(cmw.Endorsements)
		case "evidence", "ev":
			ind.Set(cmw.Evidence)
		case "attestation-results", "ar":
			ind.Set(cmw.AttestationResults)
		case "appraisal-policy", "ap":
			ind.Set(cmw.AppraisalPolicy)
		default:
			return nil, fmt.Errorf("recieved invalid indicator. Recieved %s. Accepted values can be found by running cmwcli create --help.", v[1])
		}

		createDataStruct[i] = data{
			label:     v[0],
			indicator: ind,
			filePath:  v[2],
		}
	}

	return createDataStruct, nil
}

func init() {
	rootCmd.AddCommand(createCmd)
}
