package modules

import (
	"fmt"
	"net"

	"go.uber.org/fx"
)

// A Fx module is a shareable Go library or package that provides self-contained functionality to a Fx application

// define a module

var Module = fx.Module("server", fx.Provide(
	fx.Provide(new(net.Listener)),
	fx.Invoke(startServer),
))

func startServer(l net.Listener) error {
	fmt.Println("starting server")
	return nil
}
