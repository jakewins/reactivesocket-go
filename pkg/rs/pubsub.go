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

func NewSubscriber(onSubscribe func(Subscription), onNext func(interface{}),
	onError func(error), onComplete func()) Subscriber {
	if onSubscribe == nil {
		panic("Cannot create an anonymous subscriber without the onSubscribe" +
			"function - the other functions are optional, but without this the " +
			"subscriber cannot work.")
	}
	if onNext == nil {
		onNext = func(v interface{}) {}
	}
	if onComplete == nil {
		onComplete = func() {}
	}
	if onError == nil {
		onError = func(e error) {
			fmt.Printf("Unhandled error in anonymous subscriber: %s", e.Error())
		}
	}
	return &anonymousSubscriber{onSubscribe, onNext, onError, onComplete}
}

type anonymousSubscriber struct {
	onSubscribe func(Subscription)
	onNext      func(interface{})
	onError     func(error)
	onComplete  func()
}

func (as *anonymousSubscriber) OnSubscribe(s Subscription) {
	as.onSubscribe(s)
}
func (as *anonymousSubscriber) OnNext(v interface{}) {
	as.onNext(v)
}
func (as *anonymousSubscriber) OnError(e error) {
	as.onError(e)
}
func (as *anonymousSubscriber) OnComplete() {
	as.onComplete()
}

func NewSubscription(request func(int), cancel func()) Subscription {
	if request == nil {
		panic("Cannot create a subscription without request function implemented.")
	}
	if cancel == nil {
		panic("Cannot create a subscription without cancel function implemented.")
	}
	return &anonymousSubscription{request, cancel}
}

type anonymousSubscription struct {
	request func(int)
	cancel  func()
}

func (as *anonymousSubscription) Request(n int) {
	as.request(n)
}
func (as *anonymousSubscription) Cancel() {
	as.cancel()
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
