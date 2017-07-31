// Copyright 2017 Gavin Bong. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.
package core

// Usage is the help blurb
var Usage = `
Usage:
  goping -c 2 -d www.google.com

Options:
  -c count    Stop after sending count ECHO_REQUEST packets.
              (OPTIONAL: Defaults to 5.)
  -d host     IPv4 FQDN or numeric address.
  -h          Show this message.
  -v          Increase verbosity.
`
