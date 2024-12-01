package relayinfo

// AddSupportedNIP appends a supported NIP number to a RelayInfo.
func (ri *T) AddSupportedNIP(n no) {
	idx, exists := ri.Nips.HasNumber(n)
	if exists {
		return
	}
	ri.Nips = append(ri.Nips, -1)
	copy(ri.Nips[idx+1:], ri.Nips[idx:])
	ri.Nips[idx] = n
}

type Admission struct {
	Amount no `json:"amount"`
	Unit   st `json:"unit"`
}

type Subscription struct {
	Amount no `json:"amount"`
	Unit   st `json:"unit"`
	Period no `json:"period"`
}

type Publication []struct {
	Kinds  []no `json:"kinds"`
	Amount no   `json:"amount"`
	Unit   st   `json:"unit"`
}

// Fees defines the fee structure used for a paid relay.
type Fees struct {
	Admission    []Admission    `json:"admission,omitempty"`
	Subscription []Subscription `json:"subscription,omitempty"`
	Publication  []Publication  `json:"publication,omitempty"`
}
