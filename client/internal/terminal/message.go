package terminal

type Event struct {
	Action string `json:"action"`
	Id     string `json:"id"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
	Data   string `json:"data"`
}

type Key struct {
	Action string `json:"action"`
	Id     string `json:"id"`
	Data   string `json:"data"`
}
