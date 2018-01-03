package configs

type AnnotateConf struct {
	Listener string `yaml:"listener"`
	Static   string `yaml:"static"`
}

func GetAnnotateConf(conf AnnotateConf) AnnotateConf {
	var zeroString string

	out := AnnotateConf{
		Listener: "127.0.0.1:8765",
		Static:   conf.Static,
	}
	if conf.Listener != zeroString {
		out.Listener = conf.Listener
	}

	return out
}
