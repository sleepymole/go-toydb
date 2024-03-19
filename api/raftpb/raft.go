package raftpb

type Eventer interface {
	GetEvent() isMessage_Event
}

func (h *Heartbeat) GetEvent() isMessage_Event {
	return &Message_Heartbeat{Heartbeat: h}
}

func (c *ConfirmLeader) GetEvent() isMessage_Event {
	return &Message_ConfirmLeader{ConfirmLeader: c}
}

func (s *SolicitVote) GetEvent() isMessage_Event {
	return &Message_SolicitVote{SolicitVote: s}
}

func (g *GrantVote) GetEvent() isMessage_Event {
	return &Message_GrantVote{GrantVote: g}
}

func (a *AppendEntries) GetEvent() isMessage_Event {
	return &Message_AppendEntries{AppendEntries: a}
}

func (a *AcceptEntries) GetEvent() isMessage_Event {
	return &Message_AcceptEntries{AcceptEntries: a}
}

func (r *RejectEntries) GetEvent() isMessage_Event {
	return &Message_RejectEntries{RejectEntries: r}
}

func (c *ClientRequest) GetEvent() isMessage_Event {
	return &Message_ClientRequest{ClientRequest: c}
}

func (c *ClientResponse) GetEvent() isMessage_Event {
	return &Message_ClientResponse{ClientResponse: c}
}
