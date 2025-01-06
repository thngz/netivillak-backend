package message

type SuccessResponse struct {
	Data SuccessMessage `json:"data"`
	Kind string         `json:"kind"`
}

type SuccessMessage struct {
	Message interface{} `json:"message"`
}

func InitSuccessResponse(msg interface{}) *SuccessResponse {
	return &SuccessResponse{
		Data: SuccessMessage{
			Message: msg,
		},
		Kind: "ok",
	}
}

type FailureResponse struct {
	Data FailureMessage `json:"data"`
	Kind string         `json:"kind"`
}

type FailureMessage struct {
	Err string `json:"err"`
}

func InitFailureResponse(err string) *FailureResponse {
	return &FailureResponse{
		Data: FailureMessage{
			Err: err,
		},
		Kind: "err",
	}
}
