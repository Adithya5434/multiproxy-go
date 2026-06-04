package protocols

type Proxy struct {
	Username string
	Password string
	Protected bool
}

type MCProxy struct {
	MCHost string
	MCPort int
}

type WebProxy struct {
    Host string
    Port int
}