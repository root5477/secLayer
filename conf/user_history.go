package conf

import (
	"sync"
)

type UserBuyHistory struct {
	History map[int] int
	Locker sync.RWMutex
}

func (p *UserBuyHistory) GetProductBuyCount(productId int) int {
	p.Locker.RLock()
	defer p.Locker.RUnlock()
	count, _ := p.History[productId]
	return count
}

func (p *UserBuyHistory) Add(productId, count int)  {
	p.Locker.Lock()
	defer p.Locker.Unlock()
	cur, ok := p.History[productId]
	if !ok {
		p.History[productId] = count
	} else {
		p.History[productId] = count + cur
	}
}
