package main

type InteractiveItem interface {
	getName() string
}

type key struct {
	name string
	d    door
}

func (k key) getName() string {
	return k.name
}

func setDoor(k *key, d door) {
	k.d = d
}

type door struct {
	name   string
	r1     *room
	r2     *room
	k      *key
	isOpen bool
}

func (d *door) setKey(k key) {
	d.k = &k
}

func (d door) getName() string {
	return d.name
}

func (d *door) openClose(open bool) {
	d.isOpen = open
}

func (d *door) setRoom1(r room) {
	d.r1 = &r
}

func (d *door) setRoom2(r room) {
	d.r2 = &r
}

type uselessItem struct {
	name string
}

func (item uselessItem) getName() string {
	return item.name
}

type backpack struct {
	name     string
	stash    []InteractiveItem
	capacity int
}

func (b backpack) getName() string {
	return b.name
}

func (b *backpack) putItem(item InteractiveItem) {
	b.stash = append(b.stash, item)
}
