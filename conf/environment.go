package conf

// environment
type EnvironmentEnum int8

const (
	ExampleEnvironmentEnum           EnvironmentEnum = 0x01
)

var SystemEnvironmentEnum = ExampleEnvironmentEnum

func GetYaml() string {
	var (
		//ConfigFile = "conf/conf_example.yaml"
		ConfigFile = "conf/conf_pro.yaml"
	)
	return ConfigFile
}

func MDBEnvironment() string {
	//return "conf/mdb_example.json"
	return "conf/mdb_pro.json"
}
