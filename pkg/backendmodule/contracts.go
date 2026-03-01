package backendmodule

type Manifest struct {
	AppID        string
	Name         string
	Description  string
	Required     bool
	Capabilities []string
}

type ReflectionDocument struct {
	AppID        string
	Name         string
	Version      string
	Summary      string
	Capabilities []ReflectionCapability
	Docs         []ReflectionDocLink
	APIs         []ReflectionAPI
	Schemas      []ReflectionSchemaRef
}

type ReflectionCapability struct {
	ID          string
	Stability   string
	Description string
}

type ReflectionDocLink struct {
	ID          string
	Title       string
	URL         string
	Path        string
	Description string
}

type ReflectionAPI struct {
	ID             string
	Method         string
	Path           string
	Summary        string
	RequestSchema  string
	ResponseSchema string
	ErrorSchema    string
	Tags           []string
}

type ReflectionSchemaRef struct {
	ID       string
	Format   string
	URI      string
	Embedded any
}
