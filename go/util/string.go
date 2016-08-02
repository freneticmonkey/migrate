package util

import "strings"

func StringInArray(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

// Params Used to easily manage strings of values
type Params struct {
	Values []string
	Sep    string
}

func (p *Params) Add(param string) {
	p.Values = append(p.Values, param)
}

func (p *Params) String() string {
	if len(p.Sep) == 0 {
		p.Sep = " "
	}
	return strings.Join(p.Values, p.Sep)
}
