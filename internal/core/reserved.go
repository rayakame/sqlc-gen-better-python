// Package core Auto-generated using python; DO NOT EDIT
// py 3.13.1 (tags/v3.13.1:0671451, Dec  3 2024, 19:06:28) [MSC v.1942 64 bit (AMD64)]
package core

func Escape(s string) string {
	if IsReserved(s) {
		return s + "_"
	}
	return s
}

func IsReserved(s string) bool {
	switch s {
	case "False":
		return true
	case "None":
		return true
	case "True":
		return true
	case "and":
		return true
	case "as":
		return true
	case "assert":
		return true
	case "async":
		return true
	case "await":
		return true
	case "break":
		return true
	case "class":
		return true
	case "continue":
		return true
	case "def":
		return true
	case "del":
		return true
	case "elif":
		return true
	case "else":
		return true
	case "except":
		return true
	case "finally":
		return true
	case "for":
		return true
	case "from":
		return true
	case "global":
		return true
	case "if":
		return true
	case "import":
		return true
	case "in":
		return true
	case "is":
		return true
	case "lambda":
		return true
	case "nonlocal":
		return true
	case "not":
		return true
	case "or":
		return true
	case "pass":
		return true
	case "raise":
		return true
	case "return":
		return true
	case "try":
		return true
	case "while":
		return true
	case "with":
		return true
	case "yield":
		return true
	case "id":
		return true
	default:
		return false
	}
}
