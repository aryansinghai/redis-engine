package core

var KeySpaceStat [4]map[string]int

func updateKeySpaceStat(num int, metric string, value int) {
	KeySpaceStat[num][metric] = value
}
