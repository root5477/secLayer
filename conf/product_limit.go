package conf

import "sync"

type ProductCountMgr struct {
	ProductCount map[int] int
	Locker sync.RWMutex
}

func NewProductMgr() *ProductCountMgr {
	return &ProductCountMgr{
		ProductCount:make(map[int] int, 10000),
	}
}

func (p *ProductCountMgr) Count(productId int) int {
	p.Locker.RLock()
	defer p.Locker.RUnlock()
	count, _ := p.ProductCount[productId]
	return count
}

func (p *ProductCountMgr) Add(productId, count int)  {
	p.Locker.Lock()
	defer p.Locker.Unlock()
	cur, ok := p.ProductCount[productId]
	if !ok {
		p.ProductCount[productId] = count
	} else {
		p.ProductCount[productId] = count + cur
	}
}
