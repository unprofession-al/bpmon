package dashboard

type Conf struct {
	Address string `yaml:"address"`
	Port    int    `yaml:"port"`
	Static  string `yaml:"static"`
}

func GetConf(conf Conf) Conf {
	var zeroString string
	var zeroInt int

	out := Conf{
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
