package configs

type DashboardConf struct {
	Address string `yaml:"address"`
	Port    int    `yaml:"port"`
	Static  string `yaml:"static"`
}

func GetDashboardConf(conf DashboardConf) DashboardConf {
	var zeroString string
	var zeroInt int

	out := DashboardConf{
		Address: "127.0.0.1",
		Port:    8910,
		Static:  conf.Static,
	}
	if conf.Address != zeroString {
		out.Address = conf.Address
	}
	if conf.Port != zeroInt {
		out.Port = conf.Port
	}

	return out
}
