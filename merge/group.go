package merger

type Group struct {
	Group    string   `yaml:"group"`
	Urls     []string `yaml:"urls"`
	Keywords string   `yaml:"keywords"`
	Track    []Track
}

type Channnel struct {
	Name       string `yaml:"name"`
	Url        string `yaml:"url"`
	Resolution string `yaml:"resolution"`
}
