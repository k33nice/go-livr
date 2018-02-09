package livr

func firstArg(args ...interface{}) interface{} {
	if len(args) > 0 {
		return args[0]
	}
	return nil
}
