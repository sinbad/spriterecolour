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

	RootCmd.Flags().StringVarP(&outputFile, "output", "o", "", "Output sprite file; default <input>_template.png")
	RootCmd.Flags().StringVarP(&outputTextureFile, "texture", "t", "", "File to write palette as texture; default <input>_palette.png")
	RootCmd.Flags().StringVarP(&outputParamsFile, "params", "p", "", "File to write shader params to; default none, stdout")
	RootCmd.SetUsageFunc(usageCommand)

}

func usageCommand(cmd *cobra.Command) error {
	usage := `
Usage:
  spriterecolour [options] <input file>

Options:
  -o, --output string    Output sprite file; default <input>_template.png
  -p, --params string    File to write shader params to; default none, stdout
  -t, --texture string   File to write palette as texture; default <input>_palette.png
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

	if len(outputFile) == 0 {
		outputFile = filepath.Join(filepath.Dir(infile),
			fmt.Sprintf("%s_template.png", filepath.Base(infile)))
	}

	if len(outputTextureFile) == 0 {
		outputTextureFile = filepath.Join(filepath.Dir(infile),
			fmt.Sprintf("%s_palette.png", filepath.Base(infile)))
	}

	palette, err := recolour.Generate(infile, outputFile, outputTextureFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(9)
	}

	fmt.Fprintf(os.Stderr, "Completed successfully\n")
	fmt.Fprintf(os.Stderr, "  Sprite template: %v\n", outputFile)
	fmt.Fprintf(os.Stderr, "  Palette texture: %v\n", outputTextureFile)

	if len(outputParamsFile) > 0 {
		// TODO generate shader code to file
	} else if len(palette) < 256 {
		for i, c := range palette {
			// TODO actually generate shader code?
			fmt.Printf("%d: %v\n", i, c)
		}
	}

}
