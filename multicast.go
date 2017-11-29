package main

type Multicast struct {
	conf *Conf
}

func NewMulticast(conf *Conf) (Actor, error) {
	return &Multicast{conf: conf}, nil
}

func (m *Multicast) UpdateConf(conf *Conf) error {
	m.conf = conf
	return nil
}
