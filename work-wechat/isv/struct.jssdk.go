package isv

// JSAPITicketResp is the response from GetJSAPITicket / GetAgentConfigTicket.
type JSAPITicketResp struct {
	Ticket    string `json:"ticket"`
	ExpiresIn int    `json:"expires_in"`
}
