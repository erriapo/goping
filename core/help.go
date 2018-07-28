// Copyright 2017 Gavin Chun Jin. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package core

// Usage is the help blurb
var Usage = `
Usage:
  goping www.usenix.org
  goping -c 2 8.8.4.4

Options:
  -c count    Stop after sending count ECHO_REQUEST packets. (OPTIONAL: Defaults to 5.)
  -h          Show this message.
  -I iface    Interface iface is an interface name. E.g. eth0, docker0 (OPTIONAL)
  -v          Increase verbosity.

Author: @GavinGastown3
`
