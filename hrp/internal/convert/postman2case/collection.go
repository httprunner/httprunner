package postman2case

/*
Postman Collection format reference:
https://schema.postman.com/json/collection/v2.0.0/collection.json
https://schema.postman.com/json/collection/v2.1.0/collection.json
*/

// TCollection represents the postman exported file
type TCollection struct {
	Info  TInfo   `json:"info"`
	Items []TItem `json:"item"`
}

// TInfo gives information about the collection
type TInfo struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Schema      string `json:"schema"`
}

// TItem contains the detail information of request and expected responses
// item could be defined recursively
type TItem struct {
	Items     []TItem     `json:"item"`
	Name      string      `json:"name"`
	Request   TRequest    `json:"request"`
	Responses []TResponse `json:"response"`
}

type TRequest struct {
	Method      string   `json:"method"`
	Headers     []TField `json:"header"`
	Body        TBody    `json:"body"`
	URL         TUrl     `json:"url"`
	Description string   `json:"description"`
}

type TResponse struct {
	Name            string   `json:"name"`
	OriginalRequest TRequest `json:"originalRequest"`
	Status          string   `json:"status"`
	Code            int      `json:"code"`
	Headers         []TField `json:"headers"`
	Body            string   `json:"body"`
}

type TUrl struct {
	Raw         string   `json:"raw"`
	Protocol    string   `json:"protocol"`
	Path        []string `json:"path"`
	Description string   `json:"description"`
	Query       []TField `json:"query"`
	Variable    []TField `json:"variable"`
}

type TField struct {
	Key         string `json:"key"`
	Value       string `json:"value"`
	Src         string `json:"src"`
	Description string `json:"description"`
	Type        string `json:"type"`
	Disabled    bool   `json:"disabled"`
	Enable      bool   `json:"enable"`
}

type TBody struct {
	Mode       string      `json:"mode"`
	FormData   []TField    `json:"formdata"`
	URLEncoded []TField    `json:"urlencoded"`
	Raw        string      `json:"raw"`
	Disabled   bool        `json:"disabled"`
	Options    interface{} `json:"options"`
}
