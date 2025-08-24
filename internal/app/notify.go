package app

import "github.com/codecrafters-io/redis-starter-go/pkg/ulid"

func (app *App) SubscribeKeySpace(id ulid.ID) chan Value {
	if ch, exists := app.keySpaceConsumer[id]; exists {
		return ch
	}
	ch := make(chan Value)
	app.keySpaceConsumer[id] = ch
	return ch
}

func (app *App) UnsubscribeKeySpace(id ulid.ID) {
	delete(app.keySpaceConsumer, id)
}

func (app *App) NotifyKeySpace(value Value) {
	for _, ch := range app.keySpaceConsumer {
		ch <- value
	}
}
