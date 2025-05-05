/*
Copyright © 2025 Matt Krueger <mkrueger@rstms.net>
All rights reserved.

Redistribution and use in source and binary forms, with or without
modification, are permitted provided that the following conditions are met:

 1. Redistributions of source code must retain the above copyright notice,
    this list of conditions and the following disclaimer.

 2. Redistributions in binary form must reproduce the above copyright notice,
    this list of conditions and the following disclaimer in the documentation
    and/or other materials provided with the distribution.

 3. Neither the name of the copyright holder nor the names of its contributors
    may be used to endorse or promote products derived from this software
    without specific prior written permission.

THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS"
AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE
ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE
LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR
CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF
SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS
INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN
CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE)
ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE
POSSIBILITY OF SUCH DAMAGE.
*/
package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
	"path/filepath"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list [DIR]",
	Short: "list files or maildirs",
	Long: `
Output the pathname of each compressed message file in the cur subdirectory
of the specified maildir. The default DIR is ~/Maildir

Flags:
    --recurse	    scan all maildirs rooted at DIR
    --uncompressed  output uncompressed message pathnames
    --all	    output all message pathnames
    --maildirs	    output maildirs containing selected files
`,
	Args: cobra.RangeArgs(0, 1),
	Run: func(cmd *cobra.Command, args []string) {
		cobra.CheckErr(ListFiles(args))
	},
}

func ListFiles(args []string) error {
	maildirs := viper.GetBool("maildirs")
	dirs, err := ListMaildirs(MaildirRoot(args))
	if err != nil {
		return err
	}
	for _, dir := range *dirs {
		files, err := ListMaildirFiles(dir)
		if err != nil {
			return err
		}
		if maildirs {
			if len(*files) > 0 {
				fmt.Printf("%s\n", dir)
			}
		} else {
			for _, file := range *files {
				fmt.Printf("%s\n", file)
			}
		}
	}

	return nil
}

func MaildirRoot(args []string) string {
	if len(args) > 0 {
		return args[0]
	}
	home, err := os.UserHomeDir()
	cobra.CheckErr(err)
	return filepath.Join(home, "Maildir")
}

func init() {
	rootCmd.AddCommand(listCmd)

}
