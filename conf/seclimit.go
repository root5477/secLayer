package conf

type SecLimit struct {
	count int
	currentTime int64
}

func (p *SecLimit) Count(nowTime int64) (couCount int) {
	//统计1秒内的访问次数，如果不是同一秒，将访问次数count重置为1
	if p.currentTime != nowTime {
		p.count = 1
		p.currentTime = nowTime
		couCount = p.count
		return
	}
	p.count ++
	couCount = p.count
	return
}

func (p *SecLimit) CheckNumNow(nowTime int64) int {
	if nowTime != p.currentTime {
		return 0
	}
	return p.count
}
