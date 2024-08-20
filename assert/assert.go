package assert

import "fmt"

func Assert(condition bool, arguments ...any) {
	if !condition {
		panic(messageFromMsgAndArgs(arguments...))
	}
}

func NoError(err error, arguments ...any) {
	if err != nil {
		panic(err.Error() + " " + messageFromMsgAndArgs(arguments...))
	}
}

func messageFromMsgAndArgs(msgAndArgs ...interface{}) string {
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
