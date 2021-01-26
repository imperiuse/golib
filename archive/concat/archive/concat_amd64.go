package archive

// Strings returns x+y concatenation result.
func Strings(x, y string) string

// StringsMulti - multiply strings concat
func StringsMulti(args ...string) string {
	return recursiveConcat(args)
}

func recursiveConcat(ss []string) string {
	if len(ss) > 1 {
		return Strings(ss[0], recursiveConcat(ss[1:]))
	}
	return ss[0]
}
