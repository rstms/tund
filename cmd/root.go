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
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

var rootCmd = &cobra.Command{
	Version: "0.0.7",
	Use:     "tund",
	Short:   "tun connection daemon",
	Long: `
Connect a local tun device with a tun device on a remote peer, exchanging packets via UDP
`,
	Run: func(cmd *cobra.Command, args []string) {
		tun := viper.GetInt("tunnel")
		laddr := viper.GetString("laddr")
		lport := viper.GetInt("lport")
		raddr := viper.GetString("raddr")
		rport := viper.GetInt("rport")
		verbose := viper.GetBool("verbose")
		tunnel, err := NewTunnel(tun, laddr, lport, raddr, rport, verbose)
		cobra.CheckErr(err)
		err = tunnel.Run()
		cobra.CheckErr(err)
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
func init() {
	cobra.OnInitialize(InitConfig)
	OptionString("logfile", "l", "", "log filename")
	OptionString("config", "c", "", "config file")
	OptionString("tunnel", "t", "0", "tunnel device index")
	OptionString("laddr", "a", "0.0.0.0", "local IP address")
	OptionString("lport", "p", "2000", "local UDP port")
	OptionString("raddr", "r", "", "remote IP address")
	OptionString("rport", "P", "2000", "remote UDP port")
	OptionSwitch("debug", "", "produce debug output")
	OptionSwitch("verbose", "v", "increase verbosity")
}
