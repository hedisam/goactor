package require

import (
	"fmt"
)

func Error(err error, msgAndArgs ...any) {
	if err == nil {
		msg := "expected error but got nil"
		providedMessage := messageFromMsgAndArgs(msgAndArgs...)
		if providedMessage != "" {
			msg = fmt.Sprintf("%s\nmsg: %s", msg, providedMessage)
		}
		panic(msg)
	}
}

func NoError(err error, msgAndArgs ...any) {
	if err != nil {
		msg := fmt.Sprintf("unexpected error: %v", err)
		providedMessage := messageFromMsgAndArgs(msgAndArgs...)
		if providedMessage != "" {
			msg = fmt.Sprintf("%s\nmsg: %s", msg, providedMessage)
		}
		panic(msg)
	}
}

// messageFromMsgAndArgs code from "github.com/stretchr/testify/require"
func messageFromMsgAndArgs(msgAndArgs ...any) string {
	if len(msgAndArgs) == 0 || msgAndArgs == nil {
		return ""
	}
	if len(msgAndArgs) == 1 {
		msg := msgAndArgs[0]
		if msgAsStr, ok := msg.(string); ok {
			return msgAsStr
		}
		return fmt.Sprintf("%+v", msg)
	}
	if len(msgAndArgs) > 1 {
		return fmt.Sprintf(msgAndArgs[0].(string), msgAndArgs[1:]...)
	}
	return ""
}
