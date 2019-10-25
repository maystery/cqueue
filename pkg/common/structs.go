package common

type Task struct {
	Image         string   `form:"image" json:"image"`
	Env           []string `form:"env" json:"env"`
	Cmd           []string `form:"cmd" json:"cmd" binding:"required"`
	ContainerName string   `form:"container_name" json:"container_name"`
	Type          string   `form:"type" json:"type"`
}

type ID struct {
	Id string `form:"Id" json:"Id" binding:"required"`
}
