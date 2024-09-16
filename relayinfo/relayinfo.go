package relayinfo

// AddSupportedNIP appends a supported NIP number to a RelayInfo.
func (ri *T) AddSupportedNIP(n int) {
	idx, exists := ri.Nips.HasNumber(n)
	if exists {
		return
	}
	ri.Nips = append(ri.Nips, -1)
	copy(ri.Nips[idx+1:], ri.Nips[idx:])
	ri.Nips[idx] = n
}

type Admission struct {
	Amount int    `json:"amount"`
	Unit   string `json:"unit"`
}

type Subscription struct {
	Amount int    `json:"amount"`
	Unit   string `json:"unit"`
	Period int    `json:"period"`
}

type Publication []struct {
	Kinds  []int  `json:"kinds"`
	Amount int    `json:"amount"`
	Unit   string `json:"unit"`
}

// Fees defines the fee structure used for a paid relay.
type Fees struct {
	Admission    []Admission    `json:"admission,omitempty"`
	Subscription []Subscription `json:"subscription,omitempty"`
	Publication  []Publication  `json:"publication,omitempty"`
}
