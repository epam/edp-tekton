package interceptor

type GerritEventBody struct {
	Project struct {
		Name string `json:"name"`
	} `json:"project"`
}

type GitEventBody struct {
	Repository struct {
		Name string `json:"name"`
	} `json:"repository"`
}
