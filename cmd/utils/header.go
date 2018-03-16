package utils

import "github.com/common-nighthawk/go-figure"

// PrintHeader writes the CLI header to the terminal
func PrintHeader() {
	logo := figure.NewFigure("apollo", "", true)
	logo.Print()
}
