package core

func Shutdown() {
	evalBgRewriteAOF([]string{})
}
