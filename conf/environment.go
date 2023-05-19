package conf

// environment
type EnvironmentEnum int8

const (
	ExampleEnvironmentEnum           EnvironmentEnum = 0x01
)

var SystemEnvironmentEnum = ExampleEnvironmentEnum

func GetYaml() string {
	var (
		ConfigFile = "conf/conf_example.yaml"
	)
	return ConfigFile
}

func MDBEnvironment() string {
	return "conf/mdb_example.json"
}
