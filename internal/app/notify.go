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
	close(app.keySpaceConsumer[id])
	delete(app.keySpaceConsumer, id)
}

func (app *App) NotifyKeySpace(value Value) {
	for _, ch := range app.keySpaceConsumer {
		ch <- value
	}
}

func (app *App) SubscribeBLPOPConsumer(id ulid.ID, key string) chan struct{} {
	ch := make(chan struct{})

	c := BLPOPConsumer{
		id:  id,
		key: key,
		ch:  ch,
	}

	cs := append(app.blpopConsumers[key], c)
	app.blpopConsumers[key] = cs

	return ch
}

func (app *App) UnsubscribeBLOPConsumer(id ulid.ID, key string) {
	cs := app.blpopConsumers[key]
	if len(cs) == 0 {
		return
	}

	for idx, c := range cs {
		if c.id != id {
			continue
		}
		close(c.ch)
		cs = append(cs[:idx], cs[idx+1:]...)
		break
	}

	app.blpopConsumers[key] = cs

}

func (app *App) NotifyAndPopBLPOPConsumer(key string) {
	cs := app.blpopConsumers[key]
	if len(cs) == 0 {
		return
	}
	c := cs[0]
	cs = cs[1:]
	app.blpopConsumers[key] = cs
	c.ch <- struct{}{}
	close(c.ch)
}
