package listener

import (
	"fmt"
	"net"
	"strings"

	serverstarter "github.com/lestrrat/go-server-starter/listener"
)

// Listen creates a listener from server_starter or new TCP sock
func Listen(desc string) (net.Listener, error) {
	if strings.HasPrefix(desc, "SERVER_STARTER:") {
		return listenServerStarter(strings.TrimPrefix(desc, "SERVER_STARTER:"))
	}

	return net.Listen("tcp", desc)
}

func listenServerStarter(desc string) (net.Listener, error) {
	listeners, err := serverstarter.Ports()
	if err != nil {
		return nil, err
	}

	allDescs := make([]string, len(listeners))
	for i, l := range listeners {
		var d string
		if t, ok := l.(serverstarter.TCPListener); ok {
			d = fmt.Sprintf("%s:%d", t.Addr, t.Port)
		} else if u, ok := l.(serverstarter.UnixListener); ok {
			d = u.Path
		}
		if desc == d {
			m, err := l.Listen()
			if err != nil {
				return nil, err
			}
			return m, nil
		}
		allDescs[i] = d
	}

	return nil, fmt.Errorf("no listener matches '%s'. available listeners are %s", desc, strings.Join(allDescs, ", "))
}
