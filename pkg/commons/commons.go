package commons

import (
	"strings"

	"github.com/manifoldco/promptui"
)

/** Check prefix in array of strings */
func HasPrefix(arr []string, prefix string) bool {
	for _, t := range arr {
		if strings.HasPrefix(prefix, t) {
			return true
		}
	}
	return false
}

// AsyncCallWithError calls a function asynchronously and returns the result and error through a channel.
func AsyncCall[T any](fn func() (T, error)) <-chan struct {
	Result T
	Err    error
} {
	// Create a channel to return both result and error
	resultCh := make(chan struct {
		Result T
		Err    error
	})

	// Run the function in a new goroutine
	go func() {
		defer close(resultCh) // Close the channel when the function is done
		result, err := fn()   // Call the function
		resultCh <- struct {
			Result T
			Err    error
		}{result, err} // Send the result and error to the channel
	}()

	return resultCh
}

func BuildConfirmationPrompt(question string, confirm bool) bool {
	if !confirm {
		prompt := promptui.Prompt{
			Label:     question,
			IsConfirm: true,
		}

		result, err := prompt.Run()
		if err != nil {
			return false
		}
		if result == "y" {
			return true
		}
	}
	return confirm
}
