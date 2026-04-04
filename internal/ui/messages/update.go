package messages

func (m *Model) AddUserMsg(text string) {
	m.messages = append(m.messages, "❯ "+text)
}

func (m *Model) AddAgentStream(text string) {
	m.messages = append(m.messages, text)
}
