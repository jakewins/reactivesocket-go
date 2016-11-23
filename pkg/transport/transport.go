package transport

type Server interface {
	// Runs the accept loop for this server, returns when the server is shut down.
	Serve() error
	// Signal the accept loop to shut down
	Shutdown()
	// Block until the server shuts down
	AwaitShutdown()
}
