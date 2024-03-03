package assert

import "fmt"

func True(condition bool, msgAndArgs ...any) {
	if !condition {
		const errMsg = "assertion true failed"
		if len(msgAndArgs) == 0 {
			panic(errMsg)
		} else {
			msg, ok := msgAndArgs[0].(string)
			if ok {
				panic(errMsg + "," + fmt.Sprintf(msg, msgAndArgs[1:]...))
			} else {
				panic(errMsg + "," + fmt.Sprint(msgAndArgs...))
			}
		}
	}
}
