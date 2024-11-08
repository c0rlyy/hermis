package broker

import (
	"fmt"
	"slices"
	"sync"
)

type Topic = string

type Que = []Event

type Event struct {
	Message string
	Type    string
}

// the idea is to make server lisen to all topics, and clients to lisen to specific events
// by checking Topic, and if its their userId
type EventNotification struct {
	Topic string
	Event Event
	Type  string
}

// TODO make the event lisner work
// event que to lisen as a subscireber for every channel
// notifications about events might not go trough in odrder they were added
// when there is a lot of them and they are not actibvly consumed
// TODO maybe introduce observer pattern in go intead of eventChannels
// TODO maybe return pointer to que so you dont have to specify which que by topic every time
type MessageBroker struct {
	EventChannels map[Topic]chan EventNotification // channels to subscribe to for events
	mu            sync.RWMutex
	Topics        map[Topic]Que
}

// error type for message broker
// TODO create specifing viarables of message broker err that are eqivaluent to errors eq, errEmptyQue, errPopping
type MessageBrokerErr struct {
	// broker may feel overly filled with errors
	// and while its true they they are not necessary, they do limit unpredicted behaviour
	// forcing to be more explicit in your actions
	Msg      string // the actual message
	Op       string // in what opeartion it failed, maybe usless
	Severity string // how serious the erros is eq. "warnign", "critical"
	Context  string // aditional context, like topic name or event name etc
	// Err      error  // optional field for wrapping errors, might be useless
}

func (e *MessageBrokerErr) Error() string {
	// if e.Err != nil {
	// 	return fmt.Sprintf("[%s] %s: %s - %v", e.Severity, e.Op, e.Msg, e.Err)
	// }
	return fmt.Sprintf("[%s] %s: %s, Context: %s", e.Severity, e.Op, e.Msg, e.Context)
}

func NewMessageBroker() *MessageBroker {
	mb := &MessageBroker{
		EventChannels: make(map[Topic]chan EventNotification),
		Topics:        make(map[Topic]Que),
	}
	// go mb.startBrodcasting()
	return mb
}

func NewMessageBrokerErr(msg, op, severity, context string) *MessageBrokerErr {
	return &MessageBrokerErr{
		Msg:      msg,
		Op:       op,
		Severity: severity,
		Context:  context,
	}

}

// deiced to return error since it might be usfull if i want to check wheater the topic was created,
// or it already exitsted, might be usless in most scenarios but this adds more control to the caller
func (mb *MessageBroker) AddTopic(topic Topic) error {
	mb.mu.Lock()
	defer mb.mu.Unlock()

	if _, exists := mb.Topics[topic]; exists {
		return NewMessageBrokerErr("topic already exits", "adding topic", "warrning", topic)
	}

	mb.Topics[topic] = make(Que, 0)
	mb.EventChannels[topic] = make(chan EventNotification, 10)
	return nil
}

// Curnetly the only way to close channel realeted to topic
func (mb *MessageBroker) DeleteTopic(topic Topic) error {
	mb.mu.Lock()
	defer mb.mu.Unlock()

	if _, ok := mb.Topics[topic]; !ok {
		return &MessageBrokerErr{
			Msg:      "topic was not found",
			Op:       "deleting topic",
			Severity: "critical",
			Context:  topic,
		}
	}

	delete(mb.Topics, topic)
	delete(mb.EventChannels, topic)
	close(mb.EventChannels[topic])
	return nil
}

// adds event to the queue for the specified topic.
// if topics doest exits it errors out, made this way for more less unpredicted behavoir
func (mb *MessageBroker) PushEvent(topic Topic, event Event) error {
	mb.mu.Lock()
	defer mb.mu.Unlock()

	_, exists := mb.Topics[topic]
	if !exists {
		return NewMessageBrokerErr("Topic does not exits", "pushing event", "critical", topic)
	}

	mb.Topics[topic] = append(mb.Topics[topic], event)
	ch, exists := mb.EventChannels[topic]
	if !exists {
		return NewMessageBrokerErr("channel for Topic does not exits", "pushing event", "critical", topic)
	}

	// TODO add logic for closed channels
	go func() {
		ch <- EventNotification{Topic: topic, Event: event, Type: "PUSHING"}
	}()

	return nil
}

// TODO check if the current solution does not increase memopryu footprint
// and decide whatehr coping it is worth it since it will be O(n) operaton
func (mb *MessageBroker) PopEvent(topic Topic) error {
	mb.mu.Lock()
	defer mb.mu.Unlock()

	events, exists := mb.Topics[topic]
	if !exists {
		return NewMessageBrokerErr("Topic does not exits", "popping event", "critical", topic)
	}
	if len(events) == 0 {
		return NewMessageBrokerErr("event que is empty", "popping event", "critical", topic)
	}
	ch, exists := mb.EventChannels[topic]
	if !exists {
		return NewMessageBrokerErr("channel for Topic does not exits", "poping event", "critical", topic)
	}

	// non blocking
	//TODO add logic for closed channel
	go func() {
		ch <- EventNotification{Topic: topic, Event: events[0], Type: "POPING"}
	}()
	// might not be optimal memory wise, dunno since the underlying array does not get dropped
	// by gc until their are no more refrences
	// but coping it is a more time consuming operation
	mb.Topics[topic] = events[1:]
	return nil
}

// TODO ADD AN ERROR IF NOT FOUND
// pops specific events from the queue for the specified topic, using comparison so event must == e from inside of events.
func (mb *MessageBroker) DeleteEvents(topic Topic, event Event) {
	mb.mu.Lock()
	defer mb.mu.Unlock()
	deleteFn := func(e Event) bool {
		return event == e
	}
	mb.Topics[topic] = slices.DeleteFunc(mb.Topics[topic], deleteFn)
}

// returns all events from que
func (mb *MessageBroker) GetEvents(topic Topic) Que {
	mb.mu.RLock()
	defer mb.mu.RUnlock()

	return mb.Topics[topic]
}

// gets first event found in the que mathcing the event passed in
func (mb *MessageBroker) GetEvent(topic Topic, event Event) (Event, error) {
	mb.mu.RLock()
	defer mb.mu.RUnlock()

	for _, val := range mb.Topics[topic] {
		if val == event {
			return val, nil
		}
	}
	return Event{}, NewMessageBrokerErr("no event was found in que", "get event", "warning", topic)
}

// curently the message broker only works properly with one subscriber for each topic,
// in future might intoruduce a type that would allow for lisning for events while
// also sending events to a buffer
func (mb *MessageBroker) SubscribeTopic(topic Topic) (chan EventNotification, error) {
	mb.mu.RLock()
	defer mb.mu.RUnlock()

	if _, ok := mb.Topics[topic]; !ok {
		return nil, NewMessageBrokerErr("topic was not found", "subscribing to topic", "critical", topic)
	}

	ch, ok := mb.EventChannels[topic]
	if !ok {
		return nil, NewMessageBrokerErr("channel for this topic does not exits", "subscribing to topic", "DISATROUS", topic)
	}
	return ch, nil
}
