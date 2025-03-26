import keyword
import sys
from pathlib import Path

kw_list = keyword.kwlist
path = Path(__file__).parent.parent.parent / "internal" / "core" / "reserved.go"

if __name__ == "__main__":

    with open(path, "w") as go_file:
        # Write the function header for escape
        go_file.write(f"""// Package core Auto-generated using python; DO NOT EDIT
// py {sys.version}
package core""")
        go_file.write("""
func Escape(s string) string {
    if IsReserved(s) {
        return s + "_"
    }
    return s
}

func IsReserved(s string) bool {
    switch s {
""")

        for kw in kw_list:
            go_file.write(f'    case "{kw}":\n')
            go_file.write("        return true\n")

        go_file.write("""    default:
            return false
    }
}
""")

    print("Go file 'reserved.go' has been generated.")
