package cmd

import (
	"fmt"

	"github.com/aws-cloudformation/rain/cfn/parse"
	"github.com/spf13/cobra"
)

var longDiff = false

var diffCmd = &cobra.Command{
	Use:                   "diff <from> <to>",
	Short:                 "Compare CloudFormation templates",
	Long:                  "Outputs a summary of the changes necessary to transform the CloudFormation template named <from> into the template named <to>.",
	Args:                  cobra.ExactArgs(2),
	Annotations:           templateAnnotation,
	DisableFlagsInUseLine: true,
	Run: func(cmd *cobra.Command, args []string) {
		leftFn, rightFn := args[0], args[1]

		left, err := parse.File(leftFn)
		if err != nil {
			panic(errorf(err, "unable to parse template '%s'", leftFn))
		}

		right, err := parse.File(rightFn)
		if err != nil {
			panic(errorf(err, "unable to parse template '%s'", leftFn))
		}

		fmt.Print(colouriseDiff(left.Diff(right), longDiff))
	},
}

func init() {
	diffCmd.Flags().BoolVarP(&longDiff, "long", "l", false, "Include unchanged elements in diff output")
	Rain.AddCommand(diffCmd)
}
