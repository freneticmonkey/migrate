package util

import (
	"fmt"
	"strings"
)

// SelectAction Generic function for a cli option selection prompt
func SelectAction(actionDesc string, options []string) (action string, err error) {

	for {
		fmt.Printf("%s: (%s): ", actionDesc, strings.Join(options, "/"))
		if _, err = fmt.Scanf("%s", &action); err != nil {
			if err.Error() != "unexpected newline" {
				fmt.Printf("%s\n", err)
				return "", err
			}
			action = ""
		}

		for _, opt := range options {
			if strings.ToLower(action) == strings.ToLower(opt) {
				return action, err
			}
		}
		LogWarnf("'%s' is not a valid option.", action)
	}
}
