package section

type (
	Processor struct {
		WebServer ProcessorWebServer `split_words:"true"`
	}
	ProcessorWebServer struct {
		ListenPort int `split_words:"true" default:"8080"`
	}
)
