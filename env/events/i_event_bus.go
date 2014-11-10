package events

import (
	"github.com/ottemo/foundation/env"
)

// adds listener to event handling stack
//   - event listening is patch based, "" - global listener on any event, "api.product" - will listen for app events starts with api.product.[...])
func (it *DefaultEventBus) RegisterListener(event string, listener env.F_EventListener) {
	if value, present := it.listeners[event]; present {
		it.listeners[event] = append(value, listener)
	} else {
		it.listeners[event] = []env.F_EventListener{listener}
	}
}

// generates new event, with following dispatching
func (it *DefaultEventBus) New(event string, args map[string]interface{}) {

	// loop over top level events
	// (i.e. "api.checkout.success" event will notify following listeners: "", "api", "api.checkout", "api.checkout.success")
	for charIdx, char := range event {
		if charIdx == 0 || char == '.' {
			levelEvent := event[0:charIdx]

			// processing listeners withing level if present
			if listeners, present := it.listeners[levelEvent]; present {
				for _, listener := range listeners {

					// processing listener, if it wants to stop handling - doing this
					if listener(event, args) == false {
						return
					}

				}
			}
		}
	}
}
