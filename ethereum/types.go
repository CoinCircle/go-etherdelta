package eth

import ()

type RSV struct {
	R string `json:"r"`
	S string `json:"s"`
	V int    `json:"v"`
}
