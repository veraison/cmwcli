/*
Copyright © 2026 Josh Rodriguez josh.rodriguez@arm.com
*/
package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/veraison/cmw"
	"os"
)

// createCmd represents the create command
var createCmd = NewCreateCmd()
var amendCmd = NewAmendCmd()

/*
Create a new CMW from the evidence files input.
If only one evidence file input, then create a CMW with a record.
If multiple evidence files input, then create a CMW with a collection of records.
All records have:

	media type = application/eat+cbor
	value = hash of evidence file
	indicator = 4 (evidence)
*/
func NewCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a Conceptual Message Wrapper from (CBOR only) evidence file(s)",
		Long: `Create a Conceptual Message Wrapper from (CBOR only) evidence file(s)
	For example:

	cmwcli create --data evidence.cbor --label "evidence" --cmwc_t "https://example.com" --output out.cbor`,
		Run: func(cmd *cobra.Command, args []string) {
			// Read flags
			data, err := cmd.Flags().GetStringSlice("data")
			if err != nil || data == nil || len(data) == 0 {
				fmt.Println("No data inputted")
				return
			}

			label, err := cmd.Flags().GetStringSlice("label")
			if err != nil {
				fmt.Println("Error parsing labels")
				fmt.Println(err)
				return
			}
			if len(data) > 1 && len(data) != len(label) {
				fmt.Println("Failed to create container - Non-equal amount of files and labels input.")
				fmt.Println("Given", len(data), "files, but", len(label), "labels.")
				fmt.Println("Each file must have its own label.")
				return
			}

			output, err := cmd.Flags().GetString("output")
			if err != nil || output == "" {
				fmt.Println("No output file specified")
				return
			}

			cmwc_t, err := cmd.Flags().GetString("cmwc_t")
			if err != nil {
				fmt.Println("Invalid __cmwc_t specified")
				fmt.Println(err)
				return
			}

			var cmwOutput cmw.CMW
			// Check how many files there are
			if len(data) > 1 {
				cmwOutputRef, err := cmw.NewCollection(cmwc_t)
				if err != nil {
					fmt.Println("Error creating collection")
					fmt.Println(err)
					return
				}
				cmwOutput = *cmwOutputRef

				for i, d := range data {
					file, err := os.ReadFile(d)
					if err != nil {
						fmt.Println("Error accessing cbor data in file", d)
						fmt.Println(err)
						return
					}

					newCMWItem, err := cmw.NewMonad("application/eat+cbor", file, 4)
					err = cmwOutput.AddCollectionItem(label[i], newCMWItem)
					if err != nil {
						fmt.Println("Error adding evidence from file", d, "to CMW collection")
						fmt.Println(err)
						return
					}
				}
			} else {
				file, err := os.ReadFile(data[0])
				if err != nil {
					fmt.Println("Error accessing cbor data in file", data[0])
					fmt.Println(err)
					return
				}
				cmwOutputRef, err := cmw.NewMonad("application/eat+cbor", file, 4)
				if err != nil {
					fmt.Println("Error creating CMW Record from", data[0])
					fmt.Println(err)
					return
				}
				cmwOutput = *cmwOutputRef
			}

			// Write to output file
			fmt.Println(cmwOutput)

			cmwBytes, err := cmwOutput.MarshalCBOR()
			if err != nil {
				fmt.Println("Error marshalling CMW to a CBOR file. Maybe check the input file is valid CBOR?")
				return
			}

			err = os.WriteFile(output, cmwBytes, 0666)
			if err != nil {
				fmt.Println("Error writing to file")
				return
			}
		},
	}
	cmd.Flags().StringSliceP("data", "d", nil, "Location of input data")
	cmd.Flags().StringSliceP("label", "l", nil, "Label for input data")
	cmd.Flags().StringP("output", "o", "", "Location of output file")
	cmd.Flags().StringP("cmwc_t", "c", "", "The __cmwc_t of the parent collection")
	return cmd
}

func NewAmendCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "amend",
		Short: "Amend CBOR Evidence file(s) to a CBOR CMW file",
		Long: `Amend CBOR Evidence file(s) to a CBOR CMW file
	For Example:

	cmwcli amend --cmw cmw.cbor --evidence evidence.cbor --label "evidence" [--evidence evidence2.cbor --label "evidence2"] [--output out.cbor]`,
		Run: func(cmd *cobra.Command, args []string) {
			// Read flags
			cmwLocation, err := cmd.Flags().GetString("cmw")
			if cmwLocation == nil || err != nil {
				fmt.Println("No cmw found")
				return
			}

			evidence, err := cmd.Flags().GetStringSlice("evidence")
			if len(evidence) == 0 || err != nil {
				fmt.Println("No evidence found")
				return
			}

			label, err := cmd.Flags().GetStringSlice("label")
			if len(label) == 0 || err != nil {
				fmt.Println("No evidence labels found")
				return
			}
			if len(evidence) != len(label) {
				fmt.Println("Failed to amend to cmw - Non-equal amount of evidence files and labels input.")
				fmt.Println("Given", len(evidence), "files, but", len(label), "labels.")
				fmt.Println("Each file must have its own label.")
				return
			}

			outputLocation, err := cmd.Flags().GetString("output")
			if err != nil {
				fmt.Println("Error getting output file name")
				fmt.Println(err)
				return
			}
			if output == nil {
				output = cmwLocation
			}

			// Open CMW File
			cmwFile, err := os.ReadFile(cmwLocation)
			if err != nil {
				fmt.Println("Error opening CMW file")
				fmt.Println(err)
				return
			}

			cmwInput, err := cmw.UnmarshalCBOR(cmwFile)
			if err != nil {
				fmt.Println("Error unmarshalling CMW file from CBOR")
				fmt.Println(err)
				return
			}

			// If a CMW record, convert to a Collection
			if cmwInput.GetKind() == KindMonad {
				monad = cmwInput.GetMonadValue()
				cmwInput = cmw.newCollection()
				cmwInput.AddCollectionItem("foo", monad) // What should the label of the first record be?
			}

			// Iterate over evidence files
			for i, e := range evidence {
				evidenceData, err := os.ReadFile(e)
				if err != nil {
					fmt.Println("Error accessing cbor data in file", d)
					fmt.Println(err)
					return
				}
				// Add evidence as records in collection
				cmwInput.AddCollectionItem(labels[i], evidenceData)
				newCMWItem, err := cmw.NewMonad("application/eat+cbor", file, 4)
				err = cmwInput.AddCollectionItem(label[i], newCMWItem)
				if err != nil {
					fmt.Println("Error adding evidence from file", d, "to CMW collection")
					fmt.Println(err)
					return
				}
			}
			// Write CMW to output file
			cmwOutput, err := cmwInput.MarhsalCBOR()
			if err != nil {
				fmt.Println("Error marshalling CMW to CBOR")
				fmt.Println(err)
				return
			}

			err := os.WriteFile(outputLocation, cmwOutput)
			if err != nil {
				fmt.Println("Error writing CMW to CBOR file")
				fmt.Print(err)
				return
			}
		},
	}

	cmd.Flags().StringP("cmw", "c", nil, "Location of CMW file")
	cmd.Flags().StringP("evidence", "e", nil, "Location of CBOR evidence")
	cmd.Flags().StringSliceP("label", "l", nil, "Label for CBOR evidence")
	cmd.Flags().StringP("output", "o", nil, "Location of output file")
	return cmd
}

func init() {
	rootCmd.AddCommand(createCmd)
}
