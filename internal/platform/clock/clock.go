package clock

import "time"

type Source struct {
	current func() time.Time
}

func System() Source {
	return Source{current: time.Now}
}

func Fixed(value time.Time) Source {
	return Source{current: func() time.Time { return value }}
}

func (s Source) UTCNow() time.Time {
	if s.current == nil {
		return time.Now().UTC()
	}
	return s.current().UTC()
}

func (s Source) ReviewDeadline(days int) time.Time {
	return s.UTCNow().AddDate(0, 0, days)
}
