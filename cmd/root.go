// Copyright © 2016 Steve Streeting
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/sinbad/spriterecolour/recolour"

	"github.com/spf13/cobra"
)

var (
	outputFile        string
	outputParamsFile  string
	paramsAsBytes     bool
	outputTextureFile string
)

// RootCmd represents the base command when called without any subcommands
var RootCmd *cobra.Command

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

func init() {
	RootCmd = &cobra.Command{
		Use:   "spriterecolour",
		Short: "Make re-colourable sprites from original authored sprites",
		Long: `SpriteRecolour takes an input sprite, and generates a new template
sprite from that which indexes a list of hues which can be changed to recolour
it at runtime with a shader. The original hues are generated as a 1D texture and
as a list of RGBs depending on whether you want to use a texture or shader
parameters to control the recolouring.`,
		Run: rootCommand,
	}

	RootCmd.Flags().StringVarP(&outputFile, "output", "o", "", "Output sprite file; default <input>_reference.png")
	RootCmd.Flags().StringVarP(&outputTextureFile, "texture", "t", "", "File to write palette as texture; default <input>_palette.png")
	RootCmd.Flags().StringVarP(&outputParamsFile, "params", "p", "", "File to write shader params to; default none")
	RootCmd.Flags().BoolVarP(&paramsAsBytes, "byte-params", "b", false, "When using --params, write values as 0-255 instead of 0.0-1.0")
	RootCmd.SetUsageFunc(usageCommand)

}

func usageCommand(cmd *cobra.Command) error {
	usage := `
Usage:
  spriterecolour [options] <input file>

Options:
  -o, --output string    Output sprite file; default <input>_reference.png
  -t, --texture string   File to write palette as texture; default <input>_palette.png
  -p, --params string    File to write shader params to; default none
                         Mutually exclusive with --texture
  -b, --byte-params      When using --params, write values as 0-255 instead of 0.0-1.0
`
	fmt.Fprintf(os.Stderr, usage)
	return nil
}

func rootCommand(cmd *cobra.Command, args []string) {
	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "Required: input sprite image\n")
		RootCmd.Usage()
		os.Exit(1)
	}
	infile := args[0]

	if len(outputTextureFile) > 0 && len(outputParamsFile) > 0 {
		fmt.Fprintf(os.Stderr, "Error: can't request both texture and params palettes together\n")
		RootCmd.Usage()
		os.Exit(1)
	}

	baseinfile := filepath.Base(infile)
	if ext := filepath.Ext(infile); len(ext) > 0 {
		baseinfile = baseinfile[:len(baseinfile)-len(ext)]
	}

	if len(outputFile) == 0 {
		outputFile = filepath.Join(filepath.Dir(infile),
			fmt.Sprintf("%s_reference.png", baseinfile))
	}

	if len(outputTextureFile) == 0 && len(outputParamsFile) == 0 {
		outputTextureFile = filepath.Join(filepath.Dir(infile),
			fmt.Sprintf("%s_palette.png", baseinfile))
	}

	palette, err := recolour.Generate(infile, outputFile, outputTextureFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(9)
	}

	if len(outputParamsFile) > 0 {
		fp, err := os.OpenFile(outputParamsFile, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(9)
		}
		defer fp.Close()
		for _, c := range palette {
			// Params are written just as parenthesis since how you feed them in
			// depends on the language / engine you're using
			if paramsAsBytes {
				fmt.Fprintf(fp, "(%v, %v, %v, %v)\n", uint8(c.R), uint8(c.G), uint8(c.B), uint8(c.A))
			} else {
				fmt.Fprintf(fp, "(%v, %v, %v, %v)\n", float64(c.R)/255.0, float64(c.G)/255.0, float64(c.B)/255.0, float64(c.A)/255.0)
			}
		}
	}

	fmt.Fprintf(os.Stderr, "Completed successfully\n")
	fmt.Fprintf(os.Stderr, "  Sprite reference: %v\n", outputFile)
	if len(outputTextureFile) > 0 {
		fmt.Fprintf(os.Stderr, "  Palette texture: %v\n", outputTextureFile)
	}
	if len(outputParamsFile) > 0 {
		fmt.Fprintf(os.Stderr, "  Params: %v\n", outputParamsFile)
	}

}
