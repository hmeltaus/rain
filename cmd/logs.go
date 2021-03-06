package cmd

import (
	"fmt"
	"sort"

	"github.com/aws-cloudformation/rain/client/cfn"
	"github.com/aws-cloudformation/rain/console/spinner"
	"github.com/aws-cloudformation/rain/console/text"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	"github.com/spf13/cobra"
)

var timeOrder = false
var longFormat = false
var allLogs = false

var uninterestingMessages = map[string]bool{
	"Resource creation Initiated": true,
	"User Initiated":              true,
}

func printLogs(logs []*types.StackEvent) {
	for _, log := range logs {
		fmt.Printf("- %s", colouriseStatus(string(log.ResourceStatus)))

		if timeOrder {
			fmt.Print(" ")
			fmt.Print(text.Yellow(*log.LogicalResourceId))
			fmt.Print(" ")
			fmt.Print(*log.ResourceType)

		}

		if longFormat && *log.PhysicalResourceId != "" {
			fmt.Print(" ")
			fmt.Print(*log.PhysicalResourceId)
		}

		if log.ResourceStatusReason != nil {
			fmt.Print(" ")
			fmt.Print(text.White(fmt.Sprintf("%q", *log.ResourceStatusReason)))
		}

		if longFormat {
			fmt.Print(" ")
			fmt.Print(*log.Timestamp)
		}

		fmt.Println()
	}
}

var logsCmd = &cobra.Command{
	Use:                   "logs <stack> (<resource>)",
	Short:                 "Show the event log for the named stack",
	Long:                  "Shows a nicely-formatted list of the event log for the named stack, optionally limiting the results to a single resource.\n\nBy default, rain will only show log entries that contain a message, for example a failure reason. You can use flags to change this behaviour.",
	Args:                  cobra.RangeArgs(1, 2),
	Aliases:               []string{"log"},
	Annotations:           stackAnnotation,
	DisableFlagsInUseLine: true,
	Run: func(cmd *cobra.Command, args []string) {
		stackName := args[0]

		// Get logs
		spinner.Status(fmt.Sprintf("Getting logs for %s", stackName))
		logs, err := cfn.GetStackEvents(stackName)
		if err != nil {
			panic(errorf(err, "failed to get events for '%s'", stackName))
		}
		spinner.Stop()

		// Filter by resource
		if len(args) > 1 {
			newLogs := make([]*types.StackEvent, 0)

			for _, log := range logs {
				if *log.LogicalResourceId == args[1] {
					newLogs = append(newLogs, log)
				}
			}

			logs = newLogs
		}

		// Filter out uninteresting messages
		newLogs := make([]*types.StackEvent, 0)
		for _, log := range logs {
			if allLogs || (log.ResourceStatusReason != nil && !uninterestingMessages[*log.ResourceStatusReason]) {
				newLogs = append(newLogs, log)
			}
		}
		logs = newLogs

		if len(logs) == 0 {
			fmt.Println("No interesting log messages to display. To see everything, use the --all flag")
			return
		}

		// Reverse order
		for i := len(logs)/2 - 1; i >= 0; i-- {
			j := len(logs) - 1 - i
			logs[i], logs[j] = logs[j], logs[i]
		}

		if timeOrder {
			printLogs(logs)
		} else {
			// Group by resource name
			names := make([]string, 0)
			groups := make(map[string][]*types.StackEvent)
			for _, log := range logs {
				name := *log.LogicalResourceId
				if _, ok := groups[name]; !ok {
					groups[name] = make([]*types.StackEvent, 0)
					names = append(names, name)
				}

				groups[name] = append(groups[name], log)
			}
			sort.Strings(names)

			// Print by group
			for _, name := range names {
				groupLogs := groups[name]
				fmt.Printf("%s:  # %s\n", text.Yellow(name), *groupLogs[0].ResourceType)
				printLogs(groupLogs)
				fmt.Println()
			}
		}
	},
}

func init() {
	logsCmd.Flags().BoolVarP(&timeOrder, "time", "t", false, "Show results in order of time instead of grouped by resource")
	logsCmd.Flags().BoolVarP(&longFormat, "long", "l", false, "Display full details")
	logsCmd.Flags().BoolVarP(&allLogs, "all", "a", false, "Include uninteresting logs")
	Rain.AddCommand(logsCmd)
}
