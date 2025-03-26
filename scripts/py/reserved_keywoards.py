import keyword
import sys
from pathlib import Path

kw_list = keyword.kwlist
path = Path(__file__).parent.parent.parent / "internal" / "core" / "reserved.go"

class IndentWriter:

    def __init__(self, file_path: Path, *, indent_char: str = " ", indent_amount: int = 4):
        self.file_path = file_path
        self.lines: list[tuple[str, int]] = []
        self.indent_char = indent_char
        self.indent_amount = indent_amount


    def write_line(self, text: str, indent_depth: int = 0):
        self.lines.append((text + "\n", indent_depth))

    def write_blank(self):
        self.lines.append(("\n", 0))

    def write_file(self):
        with open(self.file_path, "w") as file:
            for line in self.lines:
                indent: str = (self.indent_char * self.indent_amount) * line[1]
                file.write(indent + line[0])


if __name__ == "__main__":
    writer = IndentWriter(path)

    writer.write_line(f"// Package core Auto-generated using python; DO NOT EDIT")
    writer.write_line(f"// py {sys.version}")
    writer.write_line("package core")
    writer.write_blank()

    # Write Escape function
    writer.write_line("func Escape(s string) string {")
    writer.write_line("if IsReserved(s) {", 1)
    writer.write_line("return s + \"_\"", 2)
    writer.write_line("}", 1)
    writer.write_line("return s", 1)
    writer.write_line("}")
    writer.write_blank()

    # Write IsReserved function
    writer.write_line("func IsReserved(s string) bool {")
    writer.write_line("switch s {", 1)

    for kw in kw_list:
        writer.write_line(f'case "{kw}":', 2)
        writer.write_line("return true", 3)

    writer.write_line("default:", 2)
    writer.write_line("return false", 3)
    writer.write_line("}", 1)
    writer.write_line("}")

    # Write to file
    writer.write_file()
    print(f"Go file '{path.name}' has been generated.")
