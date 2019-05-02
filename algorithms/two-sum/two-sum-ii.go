package main

import (
	"fmt"
)

type Linode struct {
	value int
	next  *Linode
}

func main() {
	s1 := []int{2, 4, 5}
	s2 := []int{5, 6, 4, 9}
	l := &Linode{}
	m := &Linode{}
	lhead := l
	mhead := m
	for _, v := range s1 {
		lhead.value = v
		lhead.next = &Linode{}
		lhead = lhead.next

	}
	lhead.next = nil
	for _, v := range s2 {
		mhead.value = v
		mhead.next = &Linode{}
		mhead = mhead.next
	}
	mhead.next = nil
	sumTwo(l, m)
	for l != nil {
		if l.next == nil {
			if l.value != 0 {
				fmt.Println(l.value)
			}
		} else {
			fmt.Println(l.value)
		}
		l = l.next
	}
}

func sumTwo(l, m *Linode) {
	c := 0
	p := l
	q := m
	var sum int
	for {
		sum = p.value + q.value + c
		fmt.Println("aa")
		c = sum / 10
		sum = sum % 10
		p.value = sum
		if p.next == nil || q.next == nil {
			break
		}
		p = p.next
		q = q.next
	}
	if q.next == nil {
		return
	} else {
		p.next = q.next
		p = p.next
		for {
			if c > 0 {
				sum = p.value + c
				c = sum / 10
				sum = sum % 10
				p.value = sum
			}
			if p.next == nil {
				return
			}
			p = p.next
		}
	}
}
