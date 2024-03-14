package raftpb

type Eventer interface {
	Event() isMessage_Event
}

func (h *Heartbeat) Event() isMessage_Event {
	return &Message_Heartbeat{
		Heartbeat: h,
	}
}

func (c *ConfirmLeader) Event() isMessage_Event {
	return &Message_ConfirmLeader{
		ConfirmLeader: c,
	}
}

func (s *SolicitVote) Event() isMessage_Event {
	return &Message_SolicitVote{
		SolicitVote: s,
	}
}

func (g *GrantVote) Event() isMessage_Event {
	return &Message_GrantVote{
		GrantVote: g,
	}
}

func (a *AppendEntries) Event() isMessage_Event {
	return &Message_AppendEntries{
		AppendEntries: a,
	}
}

func (a *AcceptEntries) Event() isMessage_Event {
	return &Message_AcceptEntries{
		AcceptEntries: a,
	}
}

func (r *RejectEntries) Event() isMessage_Event {
	return &Message_RejectEntries{
		RejectEntries: r,
	}
}

func (c *ClientRequest) Event() isMessage_Event {
	return &Message_ClientRequest{
		ClientRequest: c,
	}
}

func (c *ClientResponse) Event() isMessage_Event {
	return &Message_ClientResponse{
		ClientResponse: c,
	}
}
