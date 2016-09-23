package reactive

// I'm not sure if replicating the reactive streams API
// out like this is a good idea; however, I can't find an "official"
// io.reactivestreams interface definition. This is based on the Java one

// A Publisher is a provider of a potentially unbounded number of sequenced
// elements, publishing them according to the demand received from it's
// Subscriber.
//
// A Publisher can serve multiple Subscribers, dynamically subscribed at
// various points in time.
type Publisher interface {
	// Request the Publisher to start streaming data.
	// This is a "factory method", it can be called multiple times, Each time
	// starting a new Subscription
	Subscribe(s Subscriber)
}

// Will receive a call to OnSubscribe once after being passed to Publisher#Subscribe,
// the Subscription provided lets the Subscriber request elements from the Publisher.
type Subscriber interface {
	// TODO: Wondering if it makes sense to break this into smaller interfaces,
	//       having "Subscriber" be the combination of them. That might make it
	//       easier to subscribe to a publisher without implementing all the things..

	OnSubscribe(s Subscription)
	OnNext(v interface{})
	OnError(e error)
	OnComplete()
}

// Represents a one-to-one lifecycle of a Subscriber subscribing to a Publisher
type Subscription interface {
	Request(n int)
	Cancel()
}
