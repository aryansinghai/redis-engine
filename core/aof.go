package core

import (
	"fmt"
	"log"
	"os"
	"server/config"
	"strings"
)

func DumpAllAOF() {
	// TODO: Implement this
	fp, err := os.OpenFile(config.Config.AOF_FILE, os.O_CREATE|os.O_WRONLY, os.ModeAppend)
	if err != nil {
		log.Fatalf("Failed to open AOF file: %v", err)
		return
	}
	log.Println("Dumping AOF to", config.Config.AOF_FILE)

	for k, obj := range store {
		dumpKey(fp, k, obj)
	}
	log.Println("AOF dumped to complete ", config.Config.AOF_FILE)
}

func dumpKey(fp *os.File, k string, obj *Obj) {
	cmd := fmt.Sprintf("SET %s %s", k, obj.Value)
	tokens := strings.Split(cmd, " ")
	fp.Write(Encode(tokens, false))
}
