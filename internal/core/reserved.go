// Package core Auto-generated using python; DO NOT EDIT
// py 3.13.1 (tags/v3.13.1:0671451, Dec  3 2024, 19:06:28) [MSC v.1942 64 bit (AMD64)]
package core

func Escape(s string) string {
	if IsReserved(s) {
		return s + "_"
	}
	return s
}
