package message

type Role = string

const UserRole Role = "user"
const ModelRole Role = "model"

// user text, agent text, agent thought, agent tool, user tool
type Message interface {
	GetRole() Role
}

type UserText struct {
	Text string
}

func (m UserText) GetRole() Role {
	return UserRole
}

type ModelText struct {
	Text string
}

func (m ModelText) GetRole() Role {
	return ModelRole
}

type ModelThought struct {
	Text string
}

func (m ModelThought) GetRole() Role {
	return ModelRole
}

type ModelToolRequest struct {
	Name      string
	Args      map[string]any
	Id        string
	Signature string
}

func (m ModelToolRequest) GetRole() Role {
	return ModelRole
}

type UserToolResult struct {
	Name     string
	Response any
}

func (m UserToolResult) GetRole() Role {
	return UserRole
}
