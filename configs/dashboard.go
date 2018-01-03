package configs

type DashboardConf struct {
	Listener string `yaml:"listener"`
	Static   string `yaml:"static"`
}

func GetDashboardConf(conf DashboardConf) DashboardConf {
	var zeroString string

	out := DashboardConf{
		Listener: "127.0.0.1:8910",
		Static:   conf.Static,
	}
	if conf.Listener != zeroString {
		out.Listener = conf.Listener
	}

	return out
}
