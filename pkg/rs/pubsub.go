package rs

import (
	"fmt"
)

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

// Don't really like below, tried to build a DSL that'd let you
// specify named functions to create anonymous streams easily.
// Not super happy with result.

type SubscriberParts struct {
	OnSubscribe func(Subscription)
	OnNext func(interface{})
	OnError func(error)
	OnComplete func()
}
// Fills in any nil functions
func (s *SubscriberParts) Build() Subscriber {
	if s.OnError == nil {
		s.OnError = func(e error) {
			fmt.Printf("Unhandled error: %s", e.Error())
		}
	}
	if s.OnNext == nil {
		s.OnNext = func(v interface{}) {}
	}
	if s.OnComplete == nil {
		s.OnComplete = func() {}
	}
	return &assembledSubscriber{s}
}
type assembledSubscriber struct {
	parts *SubscriberParts
}
func (as *assembledSubscriber) OnSubscribe(s Subscription) {
	as.parts.OnSubscribe(s)
}
func (as *assembledSubscriber) OnNext(v interface{}) {
	as.parts.OnNext(v)
}
func (as *assembledSubscriber) OnError(e error) {
	as.parts.OnError(e)
}
func (as *assembledSubscriber) OnComplete() {
	as.parts.OnComplete()
}

type SubscriptionParts struct {
	Request func(int)
	Cancel func()
}
func (s *SubscriptionParts) Build() Subscription {
	return &assembledSubscription{s}
}
type assembledSubscription struct {
	parts *SubscriptionParts
}
func (as *assembledSubscription) Request(n int) {
	as.parts.Request(n)
}
func (as *assembledSubscription) Cancel() {
	as.parts.Cancel()
}

func NewPublisher(subscribe func(Subscriber)) Publisher {
	return &anonymousPublisher{subscribe}
}
type anonymousPublisher struct {
	subscribe func(Subscriber)
}
func (a *anonymousPublisher) Subscribe(s Subscriber) {
	a.subscribe(s)
}