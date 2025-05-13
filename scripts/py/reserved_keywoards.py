"""Reserved keywords script."""

from __future__ import annotations

import keyword
import sys
from pathlib import Path

kw_list = keyword.kwlist
custom_kw_list = ["id"]
path = Path(__file__).parent.parent.parent / "internal" / "core" / "reserved.go"


class IndentWriter:
    """Indent writer used for dynamically generate files."""

    def __init__(self, file_path: Path, *, indent_char: str = " ", indent_amount: int = 4) -> None:
        """Construct a new indent writer object."""
        self.file_path = file_path
        self.lines: list[tuple[str, int]] = []
        self.indent_char = indent_char
        self.indent_amount = indent_amount

    def write_line(self, text: str, indent_depth: int = 0) -> None:
        """Write a line with a new line character at the end to the buffer."""
        self.lines.append((text + "\n", indent_depth))

    def write_blank(self) -> None:
        """Write a blank empty line to the buffer."""
        self.lines.append(("\n", 0))

    def write_file(self) -> None:
        """Write content to file."""
        with self.file_path.open("w") as file:
            for line in self.lines:
                indent: str = (self.indent_char * self.indent_amount) * line[1]
                file.write(indent + line[0])


if __name__ == "__main__":
    writer = IndentWriter(path)

    writer.write_line("// Package core Auto-generated using python; DO NOT EDIT")
    writer.write_line(f"// py {sys.version}")
    writer.write_line("package core")
    writer.write_blank()

    # Write Escape function
    writer.write_line("func Escape(s string) string {")
    writer.write_line("if IsReserved(s) {", 1)
    writer.write_line('return s + "_"', 2)
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

    for kw in custom_kw_list:
        writer.write_line(f'case "{kw}":', 2)
        writer.write_line("return true", 3)

    writer.write_line("default:", 2)
    writer.write_line("return false", 3)
    writer.write_line("}", 1)
    writer.write_line("}")

    # Write to file
    writer.write_file()
    print(f"Go file '{path.name}' has been generated.")
