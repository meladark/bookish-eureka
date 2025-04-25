package hw04lrucache

type List interface {
	Len() int
	Front() *ListItem
	Back() *ListItem
	PushFront(v interface{}) *ListItem
	PushBack(v interface{}) *ListItem
	Remove(i *ListItem)
	MoveToFront(i *ListItem)
}

type ListItem struct {
	Value interface{}
	Next  *ListItem
	Prev  *ListItem
}

type list struct {
	len   int
	front *ListItem
	back  *ListItem
	// Place your code here.
}

// Back implements List.
func (l *list) Back() *ListItem {
	return l.back
}

// Front implements List.
func (l *list) Front() *ListItem {
	return l.front
}

// MoveToFront implements List.
func (l *list) MoveToFront(i *ListItem) {
	if i == l.front {
		return
	}
	l.Remove(i)
	i.Next = l.front
	i.Prev = nil
	if l.front != nil {
		l.front.Prev = i
	}
	l.front = i
	l.len++
}

// PushBack implements List.
func (l *list) PushBack(v interface{}) *ListItem {
	newList := &ListItem{Value: v, Prev: l.back, Next: nil}
	if l.back != nil {
		l.back.Next = newList
	} else {
		l.front = newList
	}
	l.back = newList
	l.len++
	return newList
}

// PushFront implements List.
func (l *list) PushFront(v interface{}) *ListItem {
	newList := &ListItem{Value: v, Next: l.front, Prev: nil}
	if l.front != nil {
		l.front.Prev = newList
	} else {
		l.back = newList
	}
	l.front = newList
	l.len++
	return newList
}

// Remove implements List.
func (l *list) Remove(i *ListItem) {
	if i.Next != nil {
		i.Next.Prev = i.Prev
		//
	} else {
		l.back = i.Prev
	}
	if i.Prev != nil {
		i.Prev.Next = i.Next
		// i.Prev = nil
	} else {
		l.front = i.Next
	}
	i.Next = nil
	i.Prev = nil

	l.len--
}

func NewList() List {
	return new(list)
}

func (l *list) Len() int {
	return l.len
}
