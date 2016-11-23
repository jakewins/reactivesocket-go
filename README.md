# ReactiveSocket-go

This is a client and server implementation of the [ReactiveSocket Protocol](http://reactivesocket.io/).

It currently supports the TCP transport.

The project is alpha-level - all major functionality is in place, the project passes every test in the Reactive Socket TCK.
However, the API is not stable and there are likely many edge cases and bugs remaining.

## Examples

- [Connect to a TCP server](pkg/transport/tcp/example_test.go#L9)
- [Start a TCP server](pkg/transport/tcp/example_test.go#L28)

## Contributing

Contributions are very welcome.
However, *please* reach out by opening a ticket before writing big changes.
Declining hard work is the worst in the world, I'm happy to work with you to
make your contribution fits the intended direction of the project.

Small bug fixes and so on are of course welcome as immediate PRs.

## API notes

The project uses a cloned Reactive Streams API as the main API for the protocol.
However, the mainline RS API is heavily focused on asynchronous programming, which is not idiomatic in Go.

Therefore: Expect the functionality (streams with application-level back pressure) to remain the same, but
the API itself may change to be idiomatic to Go.

On a similar note: If you have suggestions for how the regular [Reactive Streams API](http://www.reactive-streams.org/)
can be adapted to be idiomatic in Go, please reach out.

## License

Apache 2, see [License](LICENSE)