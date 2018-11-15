package biostar2

type API struct {
	url        string
	username   string
	password   string
	authorized bool
	sessionID  string
}

type UserAuthWrapper struct {
	User UserAuth `json:"User"`
}

type UserAuth struct {
	Username string `json:"login_id"`
	Password string `json:"password"`
}

type ResponseWrapper struct {
	Response Response `json:"Response"`
}

type Response struct {
	Code    string `json:"code"`
	Link    string `json:"link,omitempty"`
	Message string `json:"message,omitempty"`
}

type EventWrapper struct {
	Event Event `json:"Event"`
}

type Event struct {
	EventType EventType `json:"event_type_id"`
	Index     string    `json:"index"`
	Device    Device    `json:"device_id"`
	User      User      `json:"user_id,omitempty"`
}

type EventType struct {
	Code string `json:"code"`
	Name string `json:"name"`
	// Description string `json:"description"`
}

type User struct {
	ID   string `json:"user_id"`
	Name string `json:"name"`
}

type Device struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}
