package core

func Escape(s string) string {
	if IsReserved(s) {
		return s + "_"
	}
	return s
}
